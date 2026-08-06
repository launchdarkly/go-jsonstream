package jwriter

import (
	"io"
	"unicode/utf8"
)

// streamableBuffer is a byte buffer that can optionally flush its contents to an io.Writer
// whenever they reach a chunk size. The zero value is a ready-to-use in-memory buffer.
//
// Output accumulates in a plain byte slice via append operations, which the compiler can
// inline at call sites, rather than through bytes.Buffer method calls. For the same reason,
// tokenWriter appends directly to the buf field in its hottest code paths; any code that
// does so must call maybeFlush afterward so that streaming mode keeps flushing incrementally.
type streamableBuffer struct {
	buf       []byte
	dest      io.Writer
	destErr   error
	chunkSize int
}

func (b *streamableBuffer) Bytes() []byte {
	return b.buf
}

func (b *streamableBuffer) Grow(n int) {
	if cap(b.buf)-len(b.buf) < n {
		newBuf := make([]byte, len(b.buf), len(b.buf)+n)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}

func (b *streamableBuffer) SetStreamingWriter(w io.Writer, chunkSize int) {
	b.dest = w
	b.chunkSize = chunkSize
}

func (b *streamableBuffer) Flush() error {
	if b.dest != nil {
		if len(b.buf) > 0 {
			if b.destErr == nil {
				_, b.destErr = b.dest.Write(b.buf)
			}
			b.buf = b.buf[:0]
			return b.destErr
		}
	}
	return nil
}

func (b *streamableBuffer) maybeFlush() {
	if b.dest != nil && len(b.buf) >= b.chunkSize {
		_ = b.Flush()
	}
}

func (b *streamableBuffer) GetWriterError() error {
	return b.destErr
}

func (b *streamableBuffer) Write(data []byte) {
	b.buf = append(b.buf, data...)
	b.maybeFlush()
}

func (b *streamableBuffer) WriteByte(data byte) { //nolint:govet
	b.buf = append(b.buf, data)
	b.maybeFlush()
}

func (b *streamableBuffer) WriteRune(ch rune) {
	b.buf = utf8.AppendRune(b.buf, ch)
	b.maybeFlush()
}

func (b *streamableBuffer) WriteString(s string) {
	b.buf = append(b.buf, s...)
	b.maybeFlush()
}
