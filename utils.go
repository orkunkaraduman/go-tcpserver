package tcpserver

import "bufio"

// ReadBytesLimit reads bytes as bufio.Reader.ReadBytes with limit.
func ReadBytesLimit(b *bufio.Reader, delim byte, lim int) (line []byte, err error) {
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

// TrimCrLf trims CRLF at end of buf.
func TrimCrLf(buf []byte) []byte {
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
