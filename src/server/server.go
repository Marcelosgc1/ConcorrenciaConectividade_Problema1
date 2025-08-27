package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type player struct{


}




var clientMap = make(map[string]net.Conn)

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
    defer conn.Close()
    var state = 1;
    reader := bufio.NewReader(conn)

    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Cliente desconectado")
            return
        }
        switch state {
            case 1: 
                for _,v := range clientMap {
                    if v != conn {
                        v.Write([]byte(conn.RemoteAddr().String() + ":" + "\n" + msg + "\n"))
                    }    
                }
            case 2:


        }

        //fmt.Printf("%s:\nMensagem recebida: %s", conn.RemoteAddr().String(), msg)
        
        
        
    }
}
