package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
    "math/rand"
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
    inMsg chan int
}


type Storage struct{
    mutex sync.Mutex
    cards []int
}





func (im *IdManager) addPlayer(connect net.Conn) *Player {
    im.mutex.Lock()
    defer im.mutex.Unlock()

    im.count += 1
    im.clientMap[im.count] = &Player{
        connection: connect,
        id: im.count,
        inMsg: make(chan int, 1),
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


func setupPacks() []int {
    arr := []int{1, 2, 3, 4, 5}

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
    var inputData [5]any

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
            connected = common.SendRequest(encoder, 0, err)
            ownPlayer = temp
        case 1:
            temp := im.addPlayer(conn)
            connected = common.SendRequest(encoder, 0, 1, temp.id)
            ownPlayer = temp
        case 2:
            connected = common.SendRequest(encoder, 0, 0)
        case 3:
            duo := bq.queuePlayer(ownPlayer.id)
            if duo != nil {
                enemy,_ := im.alertPlayer(duo[0])
                if enemy.connection != nil {
                    enemy.inMsg <- ownPlayer.id
                    connected = common.SendRequest(encoder, 0, 0, enemy.id)
                }else {
                    connected = common.SendRequest(encoder, 0, -1)
                }
            }else {
                enemyId := <-ownPlayer.inMsg
                if enemyId == 0 {
                    fmt.Println("erro crítico")
                    connected = common.SendRequest(encoder, 0, -1)
                }else {
                    connected = common.SendRequest(encoder, 0, 0, enemyId)
                }
            }
        case 4:
            card := sto.openPack()
            connected = common.SendRequest(encoder, 0, card)
        }
        if connected != 0 {
            break
        }
        
    }
}
