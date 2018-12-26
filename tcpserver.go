// Package tcpserver is discontinued. Please visit: github.com/orkunkaraduman/go-accepter
// Package tcpserver provides TCP server implementation.
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

// A TCPServer defines parameters for running an TCP server.
type TCPServer struct {
	// TCP address to listen on.
	Addr string

	// Handler to invoke.
	Handler Handler

	// TLSConfig optionally provides a TLS configuration.
	TLSConfig *tls.Config

	// ErrorLog specifies an optional logger for errors in Handler.
	ErrorLog *log.Logger

	l       net.Listener
	conns   map[net.Conn]connContext
	connsMu sync.RWMutex
	closeCh chan struct{}
}

type connContext struct {
	conn    net.Conn
	closeCh chan struct{}
}

// Shutdown gracefully shuts down the server without interrupting any
// connections. Shutdown works by first closing all open listeners, then
// fills closeCh on Serve method of Handler, and then waiting indefinitely for
// connections to exit Serve method of Handler and then close. If the provided
// context expires before the shutdown is complete, Shutdown returns the
// context's error, otherwise it returns any error returned from closing the
// Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS
// immediately return nil. Make sure the program doesn't exit and waits
// instead for Shutdown to return.
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

// Close immediately closes all active net.Listeners and any connections.
// For a graceful shutdown, use Shutdown.
//
// Close returns any error returned from closing the Server's underlying
// Listener(s).
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

// ListenAndServe listens on the TCP network address srv.Addr and then calls
// Serve to handle requests on incoming connections. ListenAndServe returns a
// nil error after Close or Shutdown method called.
func (srv *TCPServer) ListenAndServe() error {
	addr := srv.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(l)
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls Serve to handle requests on incoming TLS connections.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
func (srv *TCPServer) ListenAndServeTLS(certFile, keyFile string) error {
	addr := srv.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.ServeTLS(l, certFile, keyFile)
}

// Serve accepts incoming connections on the Listener l, creating a new service
// goroutine for each. The service goroutines read requests and then call
// srv.Handler to reply to them. Serve returns a nil error after Close or
// Shutdown method called.
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

// ServeTLS accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines read requests and
// then call srv.Handler to reply to them. ServeTLS returns a nil error after
// Close or Shutdown method called.
//
// Additionally, files containing a certificate and matching private key for
// the server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated.. If the certificate is signed by
// a certificate authority, the certFile should be the concatenation of the
// server's certificate, any intermediates, and the CA's certificate.
func (srv *TCPServer) ServeTLS(l net.Listener, certFile, keyFile string) (err error) {
	config := srv.TLSConfig
	if config == nil {
		config = &tls.Config{}
	}
	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return
		}
	}
	tlsListener := tls.NewListener(l, config)
	return srv.Serve(tlsListener)
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
			defer func() {
				e := recover()
				if e != nil {
					errorLog.Print(e)
				}
			}()
			srv.Handler.Serve(conn, closeCh)
		}()
	}

	conn.Close()

	srv.connsMu.Lock()
	delete(srv.conns, conn)
	srv.connsMu.Unlock()
}
