# Go TCP Server library

[![GoDoc](https://godoc.org/github.com/go-tcpserver/tcpserver?status.svg)](https://godoc.org/github.com/go-tcpserver/tcpserver)

Go TCP Server library provides `tcpserver` package implements TCP server.

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
