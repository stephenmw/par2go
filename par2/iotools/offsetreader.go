package iotools

import (
	"io"
)

type offsetReader struct {
	reader io.Reader
	offset int64
}

type OffsetReader struct {
	o *offsetReader
}

func NewOffsetReader(r io.Reader, initialOffset int64) OffsetReader {
	return OffsetReader{&offsetReader{r, initialOffset}}
}

func (r OffsetReader) Read(p []byte) (n int, err error) {
	n, err = r.o.reader.Read(p)
	r.o.offset += int64(n)
	return
}

func (r OffsetReader) Offset() (n int64) {
	return r.o.offset
}
