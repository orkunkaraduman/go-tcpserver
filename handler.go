package tcpserver

import "net"

type Handler interface {
	Serve(srv *TCPServer, conn net.Conn, closeCh <-chan struct{})
}

type HandlerFunc func(srv *TCPServer, conn net.Conn, closeCh <-chan struct{})

func (f HandlerFunc) Serve(srv *TCPServer, conn net.Conn, closeCh <-chan struct{}) {
	f(srv, conn, closeCh)
}
