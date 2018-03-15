package tcpserver

import (
	"bufio"
	"errors"
)

var (
	// ErrBufferLimitExceeded is returned when specified buffer limit
	// was exceeded.
	ErrBufferLimitExceeded = errors.New("buffer limit exceeded")
)

// ReadBytesLimit reads bytes as bufio.Reader.ReadBytes with limit.
func ReadBytesLimit(b *bufio.Reader, delim byte, lim int) (line []byte, err error) {
	line = make([]byte, 0)
	for len(line) <= lim {
		var buf []byte
		buf, err = b.ReadSlice(delim)
		line = append(line, buf...)
		if err != bufio.ErrBufferFull {
			break
		}
	}
	if err == nil && len(line) > lim {
		err = ErrBufferLimitExceeded
	}
	return
}

// trimCrLf trims CRLF at end of buf. (obsolete)
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
