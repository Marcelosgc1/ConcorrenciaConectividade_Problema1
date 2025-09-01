package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/Marcelosgc1/ConcorrenciaConectividade_Problema1/common"
)


type IdManager struct{
    mutex sync.Mutex
    count int
    clientMap map[int]Player
}
type Player struct{
    connection net.Conn
    id int
}





func (im *IdManager) addPlayer(connect net.Conn) int {
    im.mutex.Lock()
    defer im.mutex.Unlock()

    im.count += 1
    im.clientMap[im.count] = Player{
        connection: connect,
        id: im.count,
    }

    return im.count
}

func (im *IdManager) loginPlayer(connect net.Conn, login int) (int, int) {
    im.mutex.Lock()
    defer im.mutex.Unlock()

    player, ok := im.clientMap[login]
    if ok && player.connection == nil{
        player.connection = connect
        im.clientMap[login] = player
    }else if player.connection != nil {
        return 0,1
    }else {
        return 0,2
    }

    return player.id, 0
}

var im = IdManager{
        count: 0,
        clientMap: map[int]Player{},
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
    var conectado = 0
    var id = 0
    var inputData [5]any

    defer func() {
        fmt.Println("teste")
        if id!=0 {
            fmt.Println("teste2")
            im.mutex.Lock()
            p:=im.clientMap[id]
            p.connection = nil
            im.clientMap[id]=p
            im.mutex.Unlock()
        }
        fmt.Println(im.clientMap[id].connection)
        conn.Close()
    }()
    
    for {
        inputData, conectado = common.ReadData(decoder, &msg)
        if conectado!=0 {
            break
        }
        switch msg.Action {
        case 0:
            login := common.ToInt(inputData[0])
            temp, err := im.loginPlayer(conn, login)
            conectado = common.SendRequest(encoder, 0, err)
            id = temp
        case 1:
            temp := im.addPlayer(conn)
            conectado = common.SendRequest(encoder, 0, 1, temp)
            id = temp
        case 2:


        }
        if conectado != 0 {
            break
        }
        
    }
}
