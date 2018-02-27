package tcpserver

import (
	"bufio"
	"errors"
	"io"
	"net"
)

var (
	DefMaxLineSize = 1 * 1024
)

var (
	errBufferLimitExceeded = errors.New("buffer limit exceeded")
	errMaxLineSizeExceeded = errors.New("max line size exceeded")
)

type TextProtocol struct {
	OnAccept    func(ctx *TextProtocolContext)
	OnQuit      func(ctx *TextProtocolContext)
	OnReadLine  func(ctx *TextProtocolContext, line string) int
	OnReadData  func(ctx *TextProtocolContext, data []byte)
	MaxLineSize int
	UserData    interface{}
}

func (prt *TextProtocol) Serve(conn net.Conn, closeCh <-chan struct{}) {
	ctx := &TextProtocolContext{
		Prt:     prt,
		Conn:    conn,
		closeCh: closeCh,
		doneCh:  make(chan struct{}, 1),
		rd:      bufio.NewReader(conn),
		wr:      bufio.NewWriter(conn),
	}
	ctx.Serve()
}

type TextProtocolContext struct {
	Prt      *TextProtocol
	Conn     net.Conn
	UserData interface{}

	closeCh <-chan struct{}
	doneCh  chan struct{}
	rd      *bufio.Reader
	wr      *bufio.Writer
}

func (ctx *TextProtocolContext) Done() {
	select {
	case ctx.doneCh <- struct{}{}:
	default:
	}
}

func (ctx *TextProtocolContext) Serve() {
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
		case <-ctx.doneCh:
			break mainloop
		default:
		}
		line, err := ReadBytesLimit(ctx.rd, '\n', maxLineSize)
		if err != nil {
			if err == errBufferLimitExceeded {
				err = errMaxLineSizeExceeded
			}
			ctx.Done()
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
				ctx.Done()
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

func (ctx *TextProtocolContext) SendLine(line string) error {
	return ctx.SendData([]byte(line + "\r\n"))
}

func (ctx *TextProtocolContext) SendData(buf []byte) error {
	nn, err := ctx.wr.Write(buf)
	if err != nil {
		ctx.Done()
		return err
	}
	if nn < len(buf) {
		ctx.Done()
		return io.ErrShortWrite
	}
	if err := ctx.wr.Flush(); err != nil {
		ctx.Done()
		return err
	}
	return nil
}
