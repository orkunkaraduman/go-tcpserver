package tcpserver

import "net"

// A Handler responds to an TCP incoming connection.
type Handler interface {
	Serve(conn net.Conn, closeCh <-chan struct{})
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as
// handlers. If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(conn net.Conn, closeCh <-chan struct{})

// Serve calls f(conn, closeCh)
func (f HandlerFunc) Serve(conn net.Conn, closeCh <-chan struct{}) {
	f(conn, closeCh)
}
