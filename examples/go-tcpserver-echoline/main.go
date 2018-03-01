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
