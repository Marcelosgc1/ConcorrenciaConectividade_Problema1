package test

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
    connectErrs int32
    createErrs  int32
    openErrs    int32
    deckErrs    int32
    queueErrs   int32
    playErrs    int32
    respErrs    int32
)

type Message struct {
	Action int             `json:"Action"`
	Data   json.RawMessage `json:"Data"`
}

func sendRequest(enc *json.Encoder, action int, payload ...int) error {
	var data []byte
	var err error
	if payload != nil {
		data, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
	}
	req := Message{
		Action: action,
		Data:   data,
	}
	if err := enc.Encode(req); err != nil {
		return fmt.Errorf("encode req: %w", err)
	}
	return nil
}

func readResponse(dec *json.Decoder) ([]interface{}, *Message, error) {
	var msg Message
	if err := dec.Decode(&msg); err != nil {
		return nil, nil, fmt.Errorf("decode msg: %w", err)
	}
	var temp []interface{}
	if len(msg.Data) > 0 {
		if err := json.Unmarshal(msg.Data, &temp); err != nil {
			return nil, nil, fmt.Errorf("unmarshal data: %w", err)
		}
	}
	return temp, &msg, nil
}

// worker: creates player (action 1), opens 3 packs (action 4 x3), sets deck (action 6), queues (action 3),
// then attempts to play cards by sending the card id as an action (best-effort: server treats numeric actions as plays).
func worker(id int, host string, port int, mode string, wg *sync.WaitGroup) {
	defer wg.Done()
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("worker %d: fail to connect: %v\n", id, err)
		atomic.AddInt32(&connectErrs, 1)
		return
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(bufio.NewReader(conn))

	workerTag := fmt.Sprintf("W%03d", id)

	var playerID int
	var cards []int

	// create player if requested
	if mode == "create-only" || mode == "create-and-open" {
		if err := sendRequest(enc, 1); err != nil {
			fmt.Printf("%s: send create request error: %v\n", workerTag, err)
			atomic.AddInt32(&createErrs, 1)
			return
		}
		resp, _, err := readResponse(dec)
		if err != nil {
			fmt.Printf("%s: read create response error: %v\n", workerTag, err)
			return
		}
		if len(resp) >= 1 {
			if v, ok := resp[0].(float64); ok {
				playerID = int(v)
			} else if v, ok := resp[0].(int); ok {
				playerID = v
			}
		}
		fmt.Printf("%s: created player id=%d\n", workerTag, playerID)
	}

	// optional login (best-effort)
	// if mode == "login-only" || mode == "create-and-open" {
	// 	if err := sendRequest(enc, 0, playerID); err != nil {
	// 		fmt.Printf("%s: send login request error: %v\n", workerTag, err)
	// 	} else {
	// 		_, _, err := readResponse(dec)
	// 		if err != nil {
	// 			fmt.Printf("%s: read login response error: %v\n", workerTag, err)
	// 		} else {
	// 			fmt.Printf("%s: login ack for id=%d\n", workerTag, playerID)
	// 		}
	// 	}
	// }

	// open 3 packs (collect card ids)
	if mode == "open-only" || mode == "create-and-open" {
		for i := 0; i < 3; i++ {
			if err := sendRequest(enc, 4); err != nil {
				fmt.Printf("%s: send open pack error: %v\n", workerTag, err)
				return
			}
			_, msg, err := readResponse(dec)
			if err != nil {
				fmt.Printf("%s: read open pack response error: %v\n", workerTag, err)
				atomic.AddInt32(&openErrs, 1)
				return
			}
			// best-effort: server encodes card id in Message.Action (observed behavior)
			card := msg.Action
			cards = append(cards, card)
			fmt.Printf("%s: opened pack #%d -> card=%d\n", workerTag, i+1, card)
			// small pause between opens
			//time.Sleep(20 * time.Millisecond)
		}
	}

	// set deck slots using action 6: payload (deckSlot, cardIndexInCards)
	// cardIndexInCards should match the appended order: 0,1,2
	if len(cards) >= 3 {
		for slot := 0; slot < 3; slot++ {
			if err := sendRequest(enc, 6, slot, slot); err != nil {
				fmt.Printf("%s: send set deck error: %v\n", workerTag, err)
				atomic.AddInt32(&deckErrs, 1)
			} else {
				// server typically responds with status action code or similar; read it
				// _, _, err := readResponse(dec)
				// if err != nil {
				// 	fmt.Printf("%s: read set deck response err: %v\n", workerTag, err)
				// }
				fmt.Printf("%s: set deck slot %d -> cardIndex %d (card=%d)\n", workerTag, slot, slot, cards[slot])
			}
		}
	}

	// show deck (action 7) -- optional, consume response
	if err := sendRequest(enc, 7); err == nil {
		_, _, _ = readResponse(dec)
	}
	//var playerID int
	// queue for battle (action 3)
	if err := sendRequest(enc, 3); err != nil {
		fmt.Printf("%s: send queue error: %v\n", workerTag, err)
		atomic.AddInt32(&deckErrs, 1)
	} else {
		fmt.Printf("%s: queued for battle\n", workerTag)
		resp, _, _ := readResponse(dec)
		playerID = int(resp[1].(float64))
	}

	// wait a bit for matchmaking and then attempt to play cards.
	// We'll do a best-effort loop: send each deck card id as an action (server treats numeric actions as plays).
	//time.Sleep(300 * time.Millisecond) // allow time to match; adjust if necessary

	for _, card := range cards {
		// send card id as action (best-effort play)
		if err := sendRequest(enc, playerID+7, card); err != nil {
			fmt.Printf("%s: send play card %d err: %v\n", workerTag, card, err)
			atomic.AddInt32(&playErrs, 1)
			continue
		}
		// read server response about the play (if any)
		_, msg, err := readResponse(dec)
		if err != nil {
			fmt.Printf("%s: read play response err: %v\n", workerTag, err)
			atomic.AddInt32(&respErrs, 1)
		} else {
			fmt.Printf("%s: played card %d -> server action=%d\n", workerTag, card, msg.Action)
		}
		//time.Sleep(80 * time.Millisecond)
		if msg.Action==3 || msg.Action==0 {
			break
		}
	}

	// short sleep to keep connection open a tiny bit
	//time.Sleep(100 * time.Millisecond)
}

func TestMain(t *testing.T) {
	host := flag.String("host", "127.0.0.1", "server host")
	port := flag.Int("port", 8080, "server port")
	concurrency := flag.Int("concurrency", 200, "number of concurrent workers")
	mode := flag.String("mode", "create-and-open", "mode: create-only | open-only | create-and-open | login-only")
	flag.Parse()

	fmt.Printf("Starting stress: %d workers -> %s:%d, mode=%s\n", *concurrency, *host, *port, *mode)

	var wg sync.WaitGroup
	wg.Add(*concurrency)
	start := time.Now()
	for i := 1; i <= *concurrency; i++ {
		go worker(i, *host, *port, *mode, &wg)
	}
	wg.Wait()
	fmt.Println("---- Stress Test Summary ----")
	fmt.Printf("Workers: %d, Duration: %v\n", *concurrency, time.Since(start))
	fmt.Printf("Connect errors: %d\n", connectErrs)
	fmt.Printf("Create errors : %d\n", createErrs)
	fmt.Printf("Open errors   : %d\n", openErrs)
	fmt.Printf("Deck errors   : %d\n", deckErrs)
	fmt.Printf("Queue errors  : %d\n", queueErrs)
	fmt.Printf("Play errors   : %d\n", playErrs)
	fmt.Printf("Resp errors   : %d\n", respErrs)
	fmt.Println("------------------------------")
}
