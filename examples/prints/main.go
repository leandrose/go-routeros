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
	fmt.Printf("the dial started")

	if err = client.Login(*username, *password); err != nil {
		panic(err.Error())
	}
	fmt.Printf("mikrotik logged\n")

	cmd1, err := client.SendCommand("/ppp/secret/print")
	if err != nil {
		panic("err")
	}
	cmd2, err := client.SendCommand("/user/print")
	if err != nil {
		panic("err")
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("cmd1: finish\n")
			wg.Done()
		}()
		for {
			select {
			case response, ok := <-cmd1:
				if !ok {
					fmt.Println("cmd1: close")
					return
				}

				switch response.Type {
				case "!re":
					if v, ok := response.Data["=name"]; ok {
						fmt.Printf("secret name = %s\n", v)
					}
				case "!done":
					fmt.Println("cmd1: done")
				case "!empty":
					fmt.Println("cmd2: empty")
				case "!trap", "!fatal":
					fmt.Printf("Error in Mikrotik: %v\n", response.Data["=message"])
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("cmd2: finish\n")
			wg.Done()
		}()
		for {
			select {
			case response, ok := <-cmd2:
				if !ok {
					fmt.Println("cmd2: close")
					return
				}

				switch response.Type {
				case "!re":
					if v, ok := response.Data["=name"]; ok {
						fmt.Printf("user name = %s\n", v)
					}
				case "!done":
					fmt.Println("cmd2: done")
				case "!empty":
					fmt.Println("cmd2: empty")
				case "!trap", "!fatal":
					fmt.Printf("Error in Mikrotik: %v\n", response.Data["=message"])
				}
			}
		}
	}()

	wg.Wait()
	fmt.Println("finish")
}
