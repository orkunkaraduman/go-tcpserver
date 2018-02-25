package tcpserver

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Addr      string
	Handler   Handler
	TLSConfig *tls.Config
	ErrorLog  *log.Logger
	UserData  interface{}

	l       net.Listener
	conns   map[net.Conn]connContext
	connsMu sync.RWMutex
	closeCh chan struct{}
}

type connContext struct {
	conn    net.Conn
	closeCh chan struct{}
}

func (srv *TCPServer) Shutdown(ctx context.Context) (err error) {
	err = srv.l.Close()
	select {
	case srv.closeCh <- struct{}{}:
	default:
	}

	srv.connsMu.RLock()
	for _, c := range srv.conns {
		select {
		case c.closeCh <- struct{}{}:
		default:
		}
	}
	srv.connsMu.RUnlock()

	for {
		select {
		case <-time.After(5 * time.Millisecond):
			srv.connsMu.RLock()
			if len(srv.conns) == 0 {
				srv.connsMu.RUnlock()
				return
			}
			srv.connsMu.RUnlock()
		case <-ctx.Done():
			srv.connsMu.RLock()
			for _, c := range srv.conns {
				c.conn.Close()
			}
			srv.connsMu.RUnlock()
			err = ctx.Err()
			return
		}
	}
}

func (srv *TCPServer) Close() (err error) {
	err = srv.l.Close()
	select {
	case srv.closeCh <- struct{}{}:
	default:
	}

	srv.connsMu.RLock()
	for _, c := range srv.conns {
		select {
		case c.closeCh <- struct{}{}:
		default:
		}
		c.conn.Close()
	}
	srv.connsMu.RUnlock()

	return
}

func (srv *TCPServer) ListenAndServe() error {
	addr := srv.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(l)
}

func (srv *TCPServer) Serve(l net.Listener) (err error) {
	srv.l = l
	srv.conns = make(map[net.Conn]connContext)
	srv.closeCh = make(chan struct{}, 1)
	defer func() {
		srv.l.Close()
	}()
	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			select {
			case <-srv.closeCh:
				err = nil
				return
			default:
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return
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

	if srv.Handler != nil {
		errorLog := srv.ErrorLog
		if errorLog == nil {
			errorLog = log.New(ioutil.Discard, "", log.LstdFlags)
		}
		func() {
			handlerConn := conn
			if srv.TLSConfig != nil {
				tlsConn := tls.Server(conn, srv.TLSConfig)
				if err := tlsConn.Handshake(); err != nil {
					errorLog.Print("tls error:", err)
					handlerConn = nil
				} else {
					handlerConn = tlsConn
				}
			}
			defer func() {
				e := recover()
				if e != nil {
					errorLog.Print("panic:", e)
				}
			}()
			if handlerConn != nil {
				srv.Handler.Serve(srv, handlerConn, closeCh)
			}
		}()
	}

	conn.Close()

	srv.connsMu.Lock()
	delete(srv.conns, conn)
	srv.connsMu.Unlock()
}
