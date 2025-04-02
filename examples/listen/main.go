package main

import (
    "context"
    "flag"
    "fmt"
    go_routeros "github.com/leandrose/go-routeros"
    "os"
    "sync"
    "time"
)

var (
    address  = flag.String("address", "127.0.0.1:8728", "RouterOS address and port")
    username = flag.String("username", "admin", "Username")
    password = flag.String("password", "admin", "Password")
    timeout  = flag.Duration("timeout", 30*time.Second, "Cancel after")
    useTLS   = flag.Bool("tls", false, "Use TLS")
)

func init() {
    if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
        panic(err)
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), *timeout)
    defer cancel()
    client, err := go_routeros.DialContext(ctx, *address)
    if err != nil {
        panic(err.Error())
    }
    client.EnableDebug()
    fmt.Printf("the dial started\n")

    if err = client.Login(*username, *password); err != nil {
        panic(err.Error())
    }
    fmt.Printf("mikrotik logged\n")

    wg := sync.WaitGroup{}

    responses, err := client.SendCommand("/interface/listen")
    if err != nil {
        panic(err.Error())
    }
    wg.Add(1)
    go func() {
        defer func() {
            wg.Done()
            fmt.Printf("responses: finish\n")
        }()
        i := 0
    loop:
        for {
            select {
            case response, ok := <-responses:
                if !ok {
                    fmt.Println("responses: close")
                    break loop
                }

                switch response.Type {
                case "!re":
                    i++
                    if v, ok := response.Data["=name"]; ok {
                        fmt.Printf("secret name = %s\n", v)
                    }
                case "!done":
                    fmt.Println("responses: done")
                    break loop
                case "!empty":
                    fmt.Println("responses: empty")
                    break loop
                case "!trap", "!fatal":
                    fmt.Printf("Error in Mikrotik: %v\n", response.Data["message"])
                    break loop
                }
            }
        }
        fmt.Printf("it was read %d secrets\n", i)
    }()

    wg.Wait()
    fmt.Println("finish")
}
