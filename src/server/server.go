package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"

	"github.com/Marcelosgc1/ConcorrenciaConectividade_Problema1/common"
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
    inMsg chan *Games
    cards []int
    deck [3]int
}


type Storage struct{
    mutex sync.Mutex
    cards []int
}


type Games struct{
    p1 int
    p2 int
    point1 int
    point2 int
    result int //1 ou 2 p/ vencedor, 3 p/ empate e 0 p/ indefinido
}

type GameHistory struct{
    mutex sync.Mutex
    allGames []*Games
}





func (im *IdManager) addPlayer(connect net.Conn) *Player {
    im.mutex.Lock()
    defer im.mutex.Unlock()

    im.count += 1
    im.clientMap[im.count] = &Player{
        connection: connect,
        id: im.count,
        inMsg: make(chan *Games),
        cards: make([]int, 0),
    }

    return im.clientMap[im.count]
}

func (im *IdManager) alertPlayer(id int) (*Player, bool) {
    im.mutex.RLock()
    defer im.mutex.RUnlock()

    x,y := im.clientMap[id]
    return x,y
}

func (im *IdManager) loginPlayer(connect net.Conn, login int) (*Player, int) {
    im.mutex.Lock()
    defer im.mutex.Unlock()
    if login>im.count {
        return nil, 2
    }
    player, ok := im.clientMap[login]
    if ok && player.connection == nil{
        player.connection = connect
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
    fmt.Println(sto.cards)
    if len(sto.cards) <= 0 {
        return 0
    }
    x := sto.cards[0]
    sto.cards = sto.cards[1:]
    return x
}

func (hg *GameHistory) newGame(p1id int, p2id int) *Games {
    hg.mutex.Lock()
    defer sto.mutex.Unlock()
    fmt.Println(sto.cards)
    newGame := Games{
        p1: p1id, p2: p2id, point1: 0, point2: 0,result: 0,
    }
    hg.allGames = append(hg.allGames, &newGame)
    return &newGame
}


func setupPacks() []int {
    arr := []int{1, 2, 3, 4, 5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26}

    for i := len(arr) - 1; i > 0; i-- {
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
    cards: setupPacks(),
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

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Erro ao aceitar conexão:", err)
            continue
        }
        fmt.Println("Novo cliente conectado!")
        
        go handleConnection(conn)
        
    }
}

func handleConnection(conn net.Conn) {
    
    var msg common.Message
    decoder := json.NewDecoder(conn)
    encoder := json.NewEncoder(conn)
    var connected = 0
    var ownPlayer *Player
    var inputData []any
    var currGame *Games

    defer func() {
        if ownPlayer!=nil {
            ownPlayer.connection = nil
        }
        //fmt.Println(im.clientMap[ownPlayer.id].connection)
        conn.Close()
    }()
    
    for {
        inputData, connected = common.ReadData(decoder, &msg)
        if connected!=0 {
            break
        }
        switch msg.Action {
        case 0:
            temp, err := im.loginPlayer(conn, common.ToInt(inputData[0]))
            connected = common.SendRequest(encoder, err)
            ownPlayer = temp
        case 1:
            temp := im.addPlayer(conn)
            connected = common.SendRequest(encoder, 1, temp.id)
            ownPlayer = temp
        case 2:
            connected = common.SendRequest(encoder, 0)
        case 3:
            duo := bq.queuePlayer(ownPlayer.id)
            if duo != nil {
                enemy,_ := im.alertPlayer(duo[0])
                if enemy.connection != nil && currGame == nil {
                    currGame = hg.newGame(ownPlayer.id, enemy.id)
                    enemy.inMsg <- currGame
                    connected = common.SendRequest(encoder, 0, enemy.id, 1)
                    
                }else {
                    connected = common.SendRequest(encoder, -1)
                }
            }else {
                currGame = <-ownPlayer.inMsg
                if currGame == nil {
                    connected = common.SendRequest(encoder, -1)
                }else {
                    connected = common.SendRequest(encoder, 0, currGame.p1, 2)
                }
                
            }

        case 4:
            card := sto.openPack()
            if card!=0 {
                ownPlayer.cards = append(ownPlayer.cards, card)
            }
            connected = common.SendRequest(encoder, card)
        case 5:
            connected = common.SendRequestList(encoder, 0, ownPlayer.cards)
        case 6:
            if len(ownPlayer.cards) <= common.ToInt(inputData[1]){
                connected = common.SendRequest(encoder, 0, -1)
                break
            }
            ownPlayer.deck[common.ToInt(inputData[0])] = ownPlayer.cards[common.ToInt(inputData[1])]
        case 7:
            connected = common.SendRequestList(encoder, 0, ownPlayer.deck[:])
        }
        
        if connected != 0 {
            break
        }
        
    }
}
