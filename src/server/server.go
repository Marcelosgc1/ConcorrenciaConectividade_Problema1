package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"

	. "github.com/Marcelosgc1/ConcorrenciaConectividade_Problema1/common"
)


type IdManager struct{
    mutex sync.RWMutex
    count int
    clientMap map[int]*Player
}

type BattleQueue struct{
    mutex sync.RWMutex
    clientQueue []int
}

type Player struct{
    connection net.Conn
    id int
    comGame chan *Games
    alert chan int
    cards []int
    deck [3]int
    enc *json.Encoder 
}


type Storage struct{
    mutex sync.Mutex
    cards []int
}



type GameHistory struct{
    mutex sync.Mutex
    allGames []*Games
}





func (im *IdManager) addPlayer(connect net.Conn, encoder *json.Encoder) *Player {
    im.mutex.Lock()
    defer im.mutex.Unlock()

    im.count += 1
    im.clientMap[im.count] = &Player{
        connection: connect,
        id: im.count,
        comGame: make(chan *Games),
        alert: make(chan int),
        cards: make([]int, 0),
        enc: encoder,
    }

    return im.clientMap[im.count]
}

func (im *IdManager) GetPlayer(id int) (*Player) {
    im.mutex.RLock()
    defer im.mutex.RUnlock()

    x := im.clientMap[id]
    return x
}

func (im *IdManager) loginPlayer(connect net.Conn, login int, encoder *json.Encoder) (*Player, int) {
    im.mutex.Lock()
    defer im.mutex.Unlock()
    if login>im.count {
        return nil, 2
    }
    player, ok := im.clientMap[login]
    if !ok {
        return nil, 2
    }else if player.connection == nil{
        player.connection = connect
        player.enc = encoder
        player.alert = make(chan int)
    }else if player.connection != nil {
        return nil,1
    }else {
        return nil,2
    }
    return im.clientMap[login], 0
}

func (bq *BattleQueue) queuePlayer(id int) []int{
    bq.mutex.Lock()
    defer bq.mutex.Unlock()

    bq.clientQueue = append(bq.clientQueue, id)

    if len(bq.clientQueue) >= 2 {
        firstTwo := bq.clientQueue[:2]
        bq.clientQueue = bq.clientQueue[2:]

        fmt.Println("Matched players:", firstTwo)
        return firstTwo
    }
    return nil
}

func (sto *Storage) openPack() int {
    sto.mutex.Lock()
    defer sto.mutex.Unlock()
    //fmt.Println(sto.cards)
    if len(sto.cards) <= 0 {
        return 0
    }
    x := sto.cards[0]
    sto.cards = sto.cards[1:]
    return x
}

func (hg *GameHistory) newGame(p1id int, p2id int) *Games {
    hg.mutex.Lock()
    defer hg.mutex.Unlock()
    //fmt.Println(sto.cards)
    newGame := Games{
        P1: p1id, P2: p2id, Point1: 0, Point2: 0,
    }
    hg.allGames = append(hg.allGames, &newGame)
    return &newGame
}


func setupPacks(N int) []int {
    arr := make([]int, N)
    for i := 0; i < N; i++ {
        arr[i] = i + 1
    }

    for i := N - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        arr[i], arr[j] = arr[j], arr[i]
    }

    fmt.Println(arr)
    return arr
}



var im = IdManager{
        count: 0,
        clientMap: map[int]*Player{},
    }

var bq = BattleQueue{
    clientQueue: make([]int, 0),
}

var sto = Storage{
    cards: setupPacks(2000),
}
    
var hg = GameHistory{
    allGames: []*Games{},
}

func main() {

    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Erro ao iniciar servidor:", err)
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Servidor iniciado na porta 8080...")

    var count int
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Erro ao aceitar conexão:", err)
            continue
        }
        fmt.Println("Novo cliente conectado!")
        
        count+=1
        go handleConnection(conn)
        println(count)
        
    }
}

func handleConnection(conn net.Conn) {
    
    var msg Message
    decoder := json.NewDecoder(conn)
    encoder := json.NewEncoder(conn)
    var connected = 0
    var ownPlayer *Player = nil
    var inputData []any
    var currGame *Games = nil

    defer func() {
        if ownPlayer!=nil {
            ownPlayer.connection = nil
        }
        if currGame!=nil {
            var id int
            if currGame.P1 == ownPlayer.id{
                id = currGame.P2
            }else {
                id = currGame.P1
            }
            im.GetPlayer(id).alert <- -99
        }
        if ownPlayer != nil {
            fmt.Println("Player",ownPlayer.id,"desconectado")
        }else {
            fmt.Println("Player sem ID desconectado")
        }
        conn.Close()
    }()
    
    for {
        inputData, connected = ReadData(decoder, &msg)
        if connected!=0 {
            break
        }
        switch msg.Action {
        case 0:
            fmt.Println("LOGIN_PAGE")
            temp, err := im.loginPlayer(conn, ToInt(inputData[0]), encoder)
            fmt.Println("LOGIN_PAGE2")
            connected = SendRequest(encoder, err)
            fmt.Println("LOGIN_PAGE3")
            ownPlayer = temp                
        case 1:
            temp := im.addPlayer(conn, encoder)
            connected = SendRequest(encoder, 1, temp.id)
            ownPlayer = temp
        case 2:
            connected = SendRequest(encoder, 0)
        case 3:
            if ownPlayer==nil {
                connected = SendRequest(encoder, -1)
                break
            }
            duo := bq.queuePlayer(ownPlayer.id)
            if duo != nil {
                enemy := im.GetPlayer(duo[0])
                if enemy.connection != nil && currGame == nil {
                    currGame = hg.newGame(ownPlayer.id, enemy.id)
                    enemy.comGame <- currGame
                    connected = SendRequest(encoder, 0, enemy.id, 1)
                }else {
                    connected = SendRequest(encoder, -1)
                }
            }else {
                currGame = <-ownPlayer.comGame
                if currGame == nil {
                    connected = SendRequest(encoder, -1)
                }else {
                    connected = SendRequest(encoder, 0, currGame.P1, 2)
                }
                
            }

        case 4:
            if ownPlayer==nil {
                connected = SendRequest(encoder, 0)
                break
            }
            card := sto.openPack()
            if card!=0 {
                ownPlayer.cards = append(ownPlayer.cards, card)
            }
            connected = SendRequest(encoder, card)
        case 5:
            connected = SendRequestList(encoder, 0, ownPlayer.cards)
        case 6:
            if ownPlayer==nil {
                SendRequest(encoder, 0, -1)
                break
            }
            if len(ownPlayer.cards) <= ToInt(inputData[1]){
                connected = SendRequest(encoder, 0, -1)
                break
            }
            ownPlayer.deck[ToInt(inputData[0])] = ownPlayer.cards[ToInt(inputData[1])]
        case 7:
            if ownPlayer == nil{
                connected = SendRequestList(encoder, 0, nil)
                break
            }
            connected = SendRequestList(encoder, 0, ownPlayer.deck[:])
        case 8:
            if currGame==nil {
                SendRequest(encoder, -1)
                break
            }
            enemy:=im.GetPlayer(currGame.P2)
            if enemy.connection==nil {
                SendRequest(encoder, -99)
                break
            }
            enemy.alert <- ToInt(inputData[0])
            result := <- ownPlayer.alert
            SendRequest(encoder, result)
            if result == 3 || result == 0 {
                currGame = nil
            }
         
        case 9:
            if currGame==nil {
                SendRequest(encoder, -1)
                break
            }
            enemy:=im.GetPlayer(currGame.P2)
            if enemy.connection==nil {
                SendRequest(encoder, -99)
                break
            }
            var result int
            enemyCard := <-ownPlayer.alert
            if enemyCard == -99 {
                SendRequest(encoder, -99)
            }else {
                if enemyCard>ToInt(inputData[0]) {
                    currGame.Point1 += 3
                    if currGame.Point1 > 3 {
                        result = 0
                    }else {
                        result = 1
                    }
                }else {
                    currGame.Point2 += 3
                    if currGame.Point2 > 3 {
                        result = 3
                    }else {
                        result = 2
                    }
                }
                enemy.alert <- 3 - result
                SendRequest(encoder, result)
                if result == 3 || result == 0 {
                    currGame = nil
                }
            }

        case 10:
            if currGame != nil {
                SendRequest(encoder, 0, currGame.P1, currGame.Point1, currGame.P2, currGame.Point2)
            }else {
                SendRequest(encoder, -1)
            }
        }
        
        if connected != 0 {
            break
        }
        
    }
}
