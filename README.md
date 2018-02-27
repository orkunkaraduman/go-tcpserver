# Go TCP Server library

[![GoDoc](https://godoc.org/github.com/go-tcpserver/tcpserver?status.svg)](https://godoc.org/github.com/go-tcpserver/tcpserver)

Go TCP Server library provides `tcpserver` package implements TCP server.

Make TCP socket programming easy. Go TCP Server library has `TCPServer` struct
seems like `Server` struct of `http` package to create TCP servers. Also offers
`TextProtocol` struct as a `Handler` to implement text based protocols.

## Examples

For more examples, examples/

### examples/go-tcpserver-simple

```go
package main

import (
	"log"
	"net"

	"github.com/go-tcpserver/tcpserver"
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

```
