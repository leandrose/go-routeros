package main

import (
	"context"
	"flag"
	"fmt"
	go_routeros "github.com/leandrose/go-routeros"
	"os"
	"strings"
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

	//responses, err := client.SendCommand("/system/resource/print", "=.proplist=uptime,cpu-load,uptime.oid,cpu-load.oid")
	//for {
	//    select {
	//    case <-responses:
	//
	//    }
	//}

	cmds := [][]string{
		[]string{"/system/resource/print"},
		[]string{"/system/resource/print", "=.proplist=uptime,cpu-load"},
		[]string{"/system/resource/print", "=disabled=yess"}, // error
	}

	i := 0
	for _, cmd := range cmds {
		i++
		responses, err := client.SendCommand(cmd[0], cmd[1:]...)
		if err != nil {
			panic(err)
		}

	loop:
		for {
			select {
			case resp, ok := <-responses:
				if !ok {
					break loop
				}
				switch resp.Type {
				case "!done":
					if !(i == 1 || i == 2) {
						panic(fmt.Sprintf("an error occurred in command: cmd=%s", strings.Join(cmd, " ")))
					}
				case "!trap":
					if i != 3 {
						panic(fmt.Sprintf("an error occurred in command: cmd=%s", strings.Join(cmd, " ")))
					} else {
						fmt.Printf("an error expected occurred in command: cmd=%s message=%s\n", strings.Join(cmd, " "), resp.Err.Error())
					}
				}
			}
		}
	}
}
