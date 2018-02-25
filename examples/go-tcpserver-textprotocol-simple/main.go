package main

import (
	"fmt"

	"github.com/go-tcpserver/tcpserver"
)

func main() {
	tp := &tcpserver.TextProtocol{
		OnAccept: func(ctx *tcpserver.TextProtocolContext) {
			ctx.SendLine("WELCOME")
		},
		OnQuit: func(ctx *tcpserver.TextProtocolContext) {
			ctx.SendLine("QUIT")
		},
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			fmt.Println(line)
			if line == "QUIT" {
				ctx.Done()
			}
			ctx.SendLine("PONG: " + line)
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
