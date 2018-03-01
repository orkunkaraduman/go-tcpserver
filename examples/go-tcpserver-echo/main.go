package main

import (
	"log"
	"net"

	"github.com/orkunkaraduman/go-tcpserver"
)

func main() {
	srv := &tcpserver.TCPServer{
		Addr: ":1234",
		Handler: tcpserver.HandlerFunc(func(conn net.Conn, closeCh <-chan struct{}) {
			for {
				var b [1]byte
				n, err := conn.Read(b[:])
				if err != nil {
					break
				}
				if n > 0 {
					n, err := conn.Write(b[:])
					if err != nil || n < 1 {
						break
					}
				}
			}
		}),
	}
	log.Fatal(srv.ListenAndServe())
}
