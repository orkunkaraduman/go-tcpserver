# Go TCP Server library

[![GoDoc](https://godoc.org/github.com/go-tcpserver/tcpserver?status.svg)](https://godoc.org/github.com/go-tcpserver/tcpserver)

Go TCP Server library provides `tcpserver` package implements TCP server.

Make TCP socket programming easy. Go TCP Server library has `TCPServer` struct
seems like `Server` struct of `http` package to create TCP servers. Also offers
`TextProtocol` struct as a `Handler` to implement text based protocols.

## Examples

For more examples, examples/

### examples/go-tcpserver-httpip

```go
package main

import (
	"fmt"
	"log"

	"github.com/go-tcpserver/tcpserver"
)

func main() {
	prt := &tcpserver.TextProtocol{
		OnAccept: func(ctx *tcpserver.TextProtocolContext) {
			ctx.SendLine("WELCOME")
		},
		OnQuit: func(ctx *tcpserver.TextProtocolContext) {
			ctx.SendLine("QUIT")
		},
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			fmt.Println(line)
			if line == "QUIT" {
				ctx.Close()
				return 0
			}
			ctx.SendLine(line)
			return 0
		},
		OnReadData: func(ctx *tcpserver.TextProtocolContext, data []byte) {

		},
	}
	srv := &tcpserver.TCPServer{
		Addr:    ":1234",
		Handler: prt,
	}
	log.Fatal(srv.ListenAndServe())
}

```
