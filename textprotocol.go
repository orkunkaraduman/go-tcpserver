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

func (tp *TextProtocol) Serve(conn net.Conn, closeCh <-chan struct{}) {
	ctx := &TextProtocolContext{
		tp:      tp,
		conn:    conn,
		closeCh: closeCh,
		doneCh:  make(chan struct{}, 1),
		rd:      bufio.NewReader(conn),
		wr:      bufio.NewWriter(conn),
	}
	ctx.Serve()
}

type TextProtocolContext struct {
	UserData interface{}

	tp      *TextProtocol
	conn    net.Conn
	closeCh <-chan struct{}
	doneCh  chan struct{}
	rd      *bufio.Reader
	wr      *bufio.Writer
}

func (ctx *TextProtocolContext) Serve() {
	maxLineSize := ctx.tp.MaxLineSize
	if maxLineSize <= 0 {
		maxLineSize = DefMaxLineSize
	}
	if ctx.tp.OnAccept != nil {
		ctx.tp.OnAccept(ctx)
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
			ctx.doneCh <- struct{}{}
			continue
		}
		line = TrimCrLf(line)
		size := ctx.tp.OnReadLine(ctx, string(line))
		if size <= 0 {
			continue
		}
		buf := make([]byte, size)
		for i := 0; i < size; {
			n, err := ctx.rd.Read(buf[i:])
			if err != nil {
				ctx.doneCh <- struct{}{}
				continue
			}
			i += n
		}
		ctx.tp.OnReadData(ctx, buf)
	}
	ctx.wr.Flush()
	if ctx.tp.OnQuit != nil {
		ctx.tp.OnQuit(ctx)
	}
}

func (ctx *TextProtocolContext) SendLine(line string) error {
	return ctx.SendData([]byte(line + "\r\n"))
}

func (ctx *TextProtocolContext) SendData(buf []byte) error {
	nn, err := ctx.wr.Write(buf)
	if err != nil {
		ctx.doneCh <- struct{}{}
		return err
	}
	if nn < len(buf) {
		ctx.doneCh <- struct{}{}
		return io.ErrShortWrite
	}
	if err := ctx.wr.Flush(); err != nil {
		ctx.doneCh <- struct{}{}
		return err
	}
	return nil
}
