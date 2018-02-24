package tcpserver

import (
	"context"
	"log"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Addr     string
	ErrorLog *log.Logger
	UserData interface{}
	OnAccept func(srv *TCPServer, conn *net.TCPConn, closeCh <-chan struct{})

	l       net.Listener
	conns   map[net.Conn]connContext
	connsMu sync.RWMutex
	closeCh chan struct{}
}

type connContext struct {
	conn    net.Conn
	closeCh chan struct{}
}

func (srv *TCPServer) Shutdown(ctx context.Context) error {
	srv.closeCh <- struct{}{}
	if err := srv.l.Close(); err != nil {
		return err
	}

	srv.connsMu.RLock()
	for _, c := range srv.conns {
		c.closeCh <- struct{}{}
	}
	srv.connsMu.RUnlock()

	for {
		select {
		case <-time.After(5 * time.Millisecond):
			srv.connsMu.RLock()
			if len(srv.conns) == 0 {
				srv.connsMu.RUnlock()
				return nil
			}
			srv.connsMu.RUnlock()
		case <-ctx.Done():
			srv.connsMu.RLock()
			for _, c := range srv.conns {
				c.conn.Close()
			}
			srv.connsMu.RUnlock()
			return ctx.Err()
		}
	}
}

func (srv *TCPServer) Close() error {
	srv.closeCh <- struct{}{}
	if err := srv.l.Close(); err != nil {
		return err
	}

	srv.connsMu.RLock()
	for _, c := range srv.conns {
		c.closeCh <- struct{}{}
		c.conn.Close()
	}
	srv.connsMu.RUnlock()

	return nil
}

func (srv *TCPServer) ListenAndServe() error {
	addr := srv.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(l)
}

func (srv *TCPServer) Serve(l net.Listener) error {
	defer l.Close()
	srv.l = l
	srv.conns = make(map[net.Conn]connContext)
	srv.closeCh = make(chan struct{}, 1)
	for {
		conn, e := l.Accept()
		if e != nil {
			select {
			case <-srv.closeCh:
				return nil
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return e
		}
		go srv.serve(conn)
	}
}

func (srv *TCPServer) serve(conn net.Conn) {
	closeCh := make(chan struct{}, 1)

	srv.connsMu.Lock()
	srv.conns[conn] = connContext{
		conn:    conn,
		closeCh: closeCh,
	}
	srv.connsMu.Unlock()

	if srv.OnAccept != nil {
		func() {
			defer func() {
				e := recover()
				if e != nil && srv.ErrorLog != nil {
					log.Print(e)
				}
			}()
			srv.OnAccept(srv, conn.(*net.TCPConn), closeCh)
		}()
	}

	conn.Close()

	srv.connsMu.Lock()
	delete(srv.conns, conn)
	srv.connsMu.Unlock()
}
