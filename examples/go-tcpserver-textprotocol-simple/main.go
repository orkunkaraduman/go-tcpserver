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
			ctx.SendLine("PONG: " + line)
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
