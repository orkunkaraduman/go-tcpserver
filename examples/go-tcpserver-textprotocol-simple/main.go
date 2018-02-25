package main

import (
	"fmt"

	"github.com/go-tcpserver/tcpserver"
)

func main() {
	tp := &tcpserver.TextProtocol{
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			fmt.Println(line)
			ctx.SendLine("OK: " + line)
			return 0
		},
		OnReadData: func(ctx *tcpserver.TextProtocolContext, data []byte) {

		},
	}
	ts := &tcpserver.TCPServer{
		Addr:    ":1234",
		Handler: tp,
	}
	ts.ListenAndServe()
}
