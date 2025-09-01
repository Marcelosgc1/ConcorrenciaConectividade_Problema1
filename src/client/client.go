package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/Marcelosgc1/ConcorrenciaConectividade_Problema1/common"
)



type Player struct{
    State int
}

var state = 0
var sendToServer *json.Encoder

var p = Player{State: 0,}



func main() {

    conn, err := net.Dial("tcp", "localhost:8080")
    
    if err != nil {
        fmt.Println("Erro ao conectar:", err)
        os.Exit(1)
    }
    defer conn.Close()
    fmt.Println("Conectado ao servidor!")

    
    var msg common.Message
    readFromServer := json.NewDecoder(conn)
    sendToServer = json.NewEncoder(conn)

    for {
        switch p.State {
        case 0:
            println("0 p/ login")
            println("1 p/ criar usuario")
            println("2 p/ sair")
            login(msg, readFromServer)
        case 1:
            fmt.Println("Nice")
        }
        if p.State == 1 {
            //break
        }
    }


}



func login(msg common.Message, dec *json.Decoder) {
    var input int
    n, err := fmt.Scanln(&input)
    if err != nil || n == 0 {
        return
    }

    switch input{
    case 0:
        fmt.Println("insira seu login: ")
        var login int
        n, err := fmt.Scanln(&login)
        if err != nil || n == 0 {
            return
        }
        common.SendRequest(sendToServer, state, 0, login)
        println("Loading...")
        common.ReadData(dec, &msg)
        switch msg.Action{
        case 0:
            fmt.Println("Login efetuado!")
            p.State = 1
        case 1:
            fmt.Println("Já existe um dispositivo logado nesta conta.")
        case 2:
            fmt.Println("Usuário não encontrado.")
        }
    case 1:
        common.SendRequest(sendToServer, state, 1)
        println("Loading...")
        temp,_ := common.ReadData(dec, &msg)
        fmt.Println("Usuário criado", temp[0])
        p.State = 1
    default:
        fmt.Println("algo")
    }
}

