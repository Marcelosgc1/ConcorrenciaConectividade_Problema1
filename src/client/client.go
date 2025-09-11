package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Marcelosgc1/ConcorrenciaConectividade_Problema1/common"
)



type Player struct{
    State int
}
type Game struct{
    active bool
    enemyId int
}

var state = 0
var sendToServer *json.Encoder

var p = Player{State: 0,}
var g = Game{active: false, enemyId: 0,}



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
        case 0: //login page
            println("0 p/ login")
            println("1 p/ criar usuario")
            println("2 p/ ping")
            println("3 p/ sair")
            login(msg, readFromServer)
        case 1:
            println("0 p/ jogar")
            println("1 p/ abrir pacote")
            println("2 p/ ver cartas")
            println("3 p/ sair")
            mainPage(msg, readFromServer)
        case 2:
            println("====GAME START====")
            p.State = 1
            gamePage(msg, readFromServer)
        }
        if p.State == -1 {
            break
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
        common.SendRequest(sendToServer, 0, login)
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
        common.SendRequest(sendToServer, 1)
        println("Loading...")
        temp,_ := common.ReadData(dec, &msg)
        fmt.Println("Usuário criado:", common.ToInt(temp[0]))
        p.State = 1
    case 2:
        start := time.Now()
        common.SendRequest(sendToServer, 2)
        common.ReadData(dec, &msg)
        fmt.Println("Ping:",time.Since(start).Microseconds(),"µs")
    case 3:
        p.State = -1
    }
}

func mainPage(msg common.Message, dec *json.Decoder) {
    var input int
    n, err := fmt.Scanln(&input)
    if err != nil || n == 0 {
        return
    }

    switch input{
    case 0:
        fmt.Println("Entrando na fila...")
        common.SendRequest(sendToServer, 3)
        temp,_:= common.ReadData(dec, &msg)
        switch msg.Action{
        case 0:
            fmt.Println("Partida encontrada contra o jogador: ", common.ToInt(temp[0]), "! Segure seus cintos...")
            p.State = 2
        case -1:
            fmt.Println("Erro crítico ocorreu")
        }
    case 1:
        common.SendRequest(sendToServer, 4)
        println("Loading...")
        common.ReadData(dec, &msg)
        if msg.Action==0 {
            fmt.Println("nova carta: Nenhuma carta disponível")
        }else {
            fmt.Println("nova carta:", msg.Action)
        }
        p.State = 1
    case 2:
        common.SendRequest(sendToServer, 5)
        println("Loading...")
        temp,_:=common.ReadData(dec, &msg)
        fmt.Println("Minhas cartas:", temp)
        
    case 3:
        p.State = -1
    }
}



func gamePage(msg common.Message, dec *json.Decoder) {
    var input int
    n, err := fmt.Scanln(&input)
    if err != nil || n == 0 {
        return
    }

    switch input{
    case 0:
        
    case 1:

    case 2:
        
    case 3:
        p.State = -1
    }
}