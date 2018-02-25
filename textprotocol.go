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
	OnReadLine  func(line string) int
	OnReadData  func(data []byte)
	MaxLineSize int

	conn    net.Conn
	closeCh <-chan struct{}

	doneCh chan struct{}
	rd     *bufio.Reader
	wr     *bufio.Writer
}

func (tp *TextProtocol) Serve(srv *TCPServer, conn net.Conn, closeCh <-chan struct{}) {
	tp.conn = conn
	tp.closeCh = closeCh
	doneCh := make(chan struct{}, 1)
	tp.doneCh = doneCh
	rd := bufio.NewReader(conn)
	tp.rd = rd
	wr := bufio.NewWriter(conn)
	tp.wr = wr
mainloop:
	for {
		select {
		case <-closeCh:
			break mainloop
		case <-doneCh:
			break mainloop
		default:
		}
		maxLineSize := tp.MaxLineSize
		if maxLineSize <= 0 {
			maxLineSize = DefMaxLineSize
		}
		line, err := readBytesLimit(rd, '\n', maxLineSize)
		if err != nil {
			doneCh <- struct{}{}
			continue
		}
		line = trimCrLf(line)
		size := tp.OnReadLine(string(line))
		if size <= 0 {
			continue
		}
		buf := make([]byte, size)
		for i := 0; i < size; {
			n, err := rd.Read(buf[i:])
			if err != nil {
				doneCh <- struct{}{}
				continue
			}
			i += n
		}
		tp.OnReadData(buf)
	}
	wr.Flush()
}

func (tp *TextProtocol) SendLine(line string) error {
	return tp.SendData([]byte(line + "\r\n"))
}

func (tp *TextProtocol) SendData(buf []byte) error {
	nn, err := tp.wr.Write(buf)
	if err != nil {
		tp.doneCh <- struct{}{}
		return err
	}
	if nn < len(buf) {
		tp.doneCh <- struct{}{}
		return io.ErrShortWrite
	}
	if err := tp.wr.Flush(); err != nil {
		tp.doneCh <- struct{}{}
		return err
	}
	return nil
}

func readBytesLimit(b *bufio.Reader, delim byte, lim int) (line []byte, err error) {
	line = make([]byte, 0)
	for len(line) <= lim {
		buf, e := b.ReadSlice(delim)
		line = append(line, buf...)
		if e != nil {
			if e == bufio.ErrBufferFull {
				continue
			}
			err = e
		}
		break
	}
	if err == nil && len(line) > lim {
		err = errBufferLimitExceeded
	}
	return
}

func trimCrLf(buf []byte) []byte {
	l := len(buf)
	if l == 0 {
		return buf
	}
	l--
	if buf[l] != '\n' {
		return buf
	}
	buf = buf[0:l]
	if l == 0 {
		return buf
	}
	l--
	if buf[l] != '\r' {
		return buf
	}
	buf = buf[0:l]
	return buf
}
