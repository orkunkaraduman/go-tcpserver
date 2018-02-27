package tcpserver

import (
	"bufio"
	"errors"
	"io"
	"net"
)

var (
	// DefMaxLineSize specifies maximum line size with delimiter if
	// TextProtocol.MaxLineSize is 0.
	DefMaxLineSize = 1 * 1024
)

var (
	errBufferLimitExceeded = errors.New("buffer limit exceeded")
	errMaxLineSizeExceeded = errors.New("max line size exceeded")
)

// TextProtocol defines parameters for Handler of text protocol.
type TextProtocol struct {
	// Accept handler. It will be called before reading line.
	OnAccept func(ctx *TextProtocolContext)

	// Quit handler. It will be called before closing.
	OnQuit func(ctx *TextProtocolContext)

	// ReadLine handler. If it returns greater then 0, context reads data from
	// connection n bytes. And after will be call OnReadData.
	OnReadLine func(ctx *TextProtocolContext, line string) (n int)

	// ReadData handler.
	OnReadData func(ctx *TextProtocolContext, data []byte)

	// MaxLineSize specifies maximum line size with delimiter.
	MaxLineSize int

	// User data to use free.
	UserData interface{}
}

// Serve implements Handler.Serve.
func (prt *TextProtocol) Serve(conn net.Conn, closeCh <-chan struct{}) {
	ctx := &TextProtocolContext{
		Prt:      prt,
		Conn:     conn,
		closeCh:  closeCh,
		closeCh2: make(chan struct{}, 1),
		rd:       bufio.NewReader(conn),
		wr:       bufio.NewWriter(conn),
	}
	ctx.serve()
}

// TextProtocolContext defines parameters for text protocol context.
type TextProtocolContext struct {
	// Pointer of TextProtocol struct handled by this context.
	Prt *TextProtocol

	// Connection handled by this context.
	Conn net.Conn

	// User data to use free.
	UserData interface{}

	closeCh  <-chan struct{}
	closeCh2 chan struct{}
	rd       *bufio.Reader
	wr       *bufio.Writer
}

// Close closes context.
func (ctx *TextProtocolContext) Close() {
	select {
	case ctx.closeCh2 <- struct{}{}:
	default:
	}
}

func (ctx *TextProtocolContext) serve() {
	maxLineSize := ctx.Prt.MaxLineSize
	if maxLineSize <= 0 {
		maxLineSize = DefMaxLineSize
	}
	if ctx.Prt.OnAccept != nil {
		ctx.Prt.OnAccept(ctx)
	}
mainloop:
	for {
		select {
		case <-ctx.closeCh:
			break mainloop
		case <-ctx.closeCh2:
			break mainloop
		default:
		}
		line, err := ReadBytesLimit(ctx.rd, '\n', maxLineSize)
		if err != nil {
			if err == errBufferLimitExceeded {
				err = errMaxLineSizeExceeded
			}
			ctx.Close()
			continue
		}
		line = TrimCrLf(line)
		size := ctx.Prt.OnReadLine(ctx, string(line))
		if size <= 0 {
			continue
		}
		buf := make([]byte, size)
		for i := 0; i < size; {
			n, err := ctx.rd.Read(buf[i:])
			if err != nil {
				ctx.Close()
				continue
			}
			i += n
		}
		ctx.Prt.OnReadData(ctx, buf)
	}
	ctx.wr.Flush()
	if ctx.Prt.OnQuit != nil {
		ctx.Prt.OnQuit(ctx)
	}
}

// SendLine writes a line to connection.
func (ctx *TextProtocolContext) SendLine(line string) error {
	return ctx.SendData([]byte(line + "\r\n"))
}

// SendData writes data to connection.
func (ctx *TextProtocolContext) SendData(buf []byte) error {
	nn, err := ctx.wr.Write(buf)
	if err != nil {
		ctx.Close()
		return err
	}
	if nn < len(buf) {
		ctx.Close()
		return io.ErrShortWrite
	}
	if err := ctx.wr.Flush(); err != nil {
		ctx.Close()
		return err
	}
	return nil
}
