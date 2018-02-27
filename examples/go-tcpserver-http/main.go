package main

import (
	"log"

	"github.com/go-tcpserver/tcpserver"
)

func main() {
	prt := &tcpserver.TextProtocol{
		OnReadLine: func(ctx *tcpserver.TextProtocolContext, line string) int {
			if line == "" {
				ctx.SendLine("HTTP/1.1 200 OK")
				ctx.SendLine("")
				ctx.SendLine(ctx.Conn.RemoteAddr().String())
				ctx.Close()
				return 0
			}
			return 0
		},
	}
	srv := &tcpserver.TCPServer{
		Addr:    ":80",
		Handler: prt,
	}
	log.Fatal(srv.ListenAndServe())
}
