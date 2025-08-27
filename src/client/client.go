package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Println("Erro ao conectar:", err)
        os.Exit(1)
    }
    defer conn.Close()
    fmt.Println("Conectado ao servidor!")

    
    go func() {
        reader := bufio.NewReader(conn)
        for {
            msg, err := reader.ReadString('\n')
            if err != nil {
                fmt.Println("Servidor desconectou")
                os.Exit(0)
            }
            fmt.Print(msg)
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        text := scanner.Text() + "\n"
        conn.Write([]byte(text))
    }
}
