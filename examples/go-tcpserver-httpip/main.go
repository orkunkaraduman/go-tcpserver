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
