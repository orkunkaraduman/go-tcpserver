# DISCONTINUED

Please visit: [github.com/orkunkaraduman/go-accepter](https://github.com/orkunkaraduman/go-accepter)

# Go TCP Server library

[![GoDoc](https://godoc.org/github.com/orkunkaraduman/go-tcpserver?status.svg)](https://godoc.org/github.com/orkunkaraduman/go-tcpserver)

Go TCP Server library provides `tcpserver` package implements TCP server.

Go TCP Server library has `TCPServer` struct
seems like `Server` struct of `http` package to create TCP servers. Also offers
`TextProtocol` struct as a `Handler` to implement text based protocols.

## Examples

For more examples, examples/

### examples/go-tcpserver-echoline

```go
package main

import (
	"fmt"
	"log"

	"github.com/orkunkaraduman/go-tcpserver"
)

func main() {
	prt := &tcpserver.TextProtocol{
		OnAccept: func(ctx *tcpserver.TextProtocolContext) {
			ctx.WriteLine("WELCOME")
		},
		OnQuit: func(ctx *tcpserver.TextProtocolContext) {
			ctx.WriteLine("QUIT")
		},
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			fmt.Println(line)
			if line == "QUIT" {
				ctx.Close()
				return 0
			}
			ctx.WriteLine(line)
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
