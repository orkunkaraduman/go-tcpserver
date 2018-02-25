package tcpserver

import "net"

type Handler interface {
	Serve(conn net.Conn, closeCh <-chan struct{})
}

type HandlerFunc func(conn net.Conn, closeCh <-chan struct{})

func (f HandlerFunc) Serve(conn net.Conn, closeCh <-chan struct{}) {
	f(conn, closeCh)
}
