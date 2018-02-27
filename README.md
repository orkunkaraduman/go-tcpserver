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
	"log"
	"strings"

	"github.com/go-tcpserver/tcpserver"
)

func main() {
	prt := &tcpserver.TextProtocol{
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			if line == "" {
				ctx.SendLine("HTTP/1.1 200 OK")
				ctx.SendLine("")
				ip := strings.SplitN(ctx.Conn.RemoteAddr().String(), ":", 2)[0]
				ctx.SendLine(ip)
				ctx.Close()
				return 0
			}
			return 0
		},
	}
	srv := &tcpserver.TCPServer{
		Addr:    ":8000",
		Handler: prt,
	}
	log.Fatal(srv.ListenAndServe())
}

```
