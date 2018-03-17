package tcpserver

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

// DefMaxLineSize specifies maximum line size with delimiter if
// TextProtocol.MaxLineSize is 0.
var DefMaxLineSize = 1 * 1024

// TextProtocol defines parameters for Handler of text based protocol.
type TextProtocol struct {
	// Accept callback. It will be called before reading line.
	OnAccept func(ctx *TextProtocolContext)

	// Quit callback. It will be called before closing.
	OnQuit func(ctx *TextProtocolContext)

	// ReadLine callback. If it returns greater then 0, context reads data from
	// connection n bytes. And after will be call OnReadData.
	OnReadLine func(ctx *TextProtocolContext, line string) (n int)

	// ReadData callback.
	OnReadData func(ctx *TextProtocolContext, buf []byte)

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
			ctx.Close()
			continue
		}
		line = bytes.TrimSuffix(line, []byte("\n"))
		line = bytes.TrimSuffix(line, []byte("\r"))
		size := ctx.Prt.OnReadLine(ctx, string(line))
		if size <= 0 {
			continue
		}
		buf := make([]byte, size)
		_, err = io.ReadFull(ctx.rd, buf)
		if err != nil {
			ctx.Close()
			continue
		}
		if ctx.Prt.OnReadData != nil {
			ctx.Prt.OnReadData(ctx, buf)
		}
	}
	ctx.wr.Flush()
	if ctx.Prt.OnQuit != nil {
		ctx.Prt.OnQuit(ctx)
	}
}

// WriteLine writes a line to connection.
func (ctx *TextProtocolContext) WriteLine(line string) error {
	return ctx.WriteData([]byte(line + "\r\n"))
}

// WriteData writes data to connection.
func (ctx *TextProtocolContext) WriteData(buf []byte) error {
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
