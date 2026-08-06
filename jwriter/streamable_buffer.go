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
//
// Token-level writes — strings, numbers, and the multi-byte Write used for keyword and
// raw tokens — reserve capacity before appending: bare append grows large slices by only
// ~1.25x, and the at-least-doubling policy in reserve keeps the number of reallocations
// (and the total bytes copied) logarithmic for large outputs. WriteByte and WriteRune
// deliberately do not reserve: a capacity check on every single-byte delimiter write
// measurably slows encoding, and a growth triggered by one is made rare by the reserves
// on the token writes around it.
//
// The chunk-size check happens after a write, never during one, so a single write larger
// than the chunk size — a raw JSON value, or a long escape-free string segment — is
// buffered whole before it is flushed.
type streamableBuffer struct {
	buf       []byte
	dest      io.Writer
	destErr   error
	chunkSize int
}

func (b *streamableBuffer) Bytes() []byte {
	return b.buf
}

// Grow ensures that the buffer has room for at least n more bytes, reallocating it if
// necessary. It panics if n is negative or if the buffer cannot grow that large.
func (b *streamableBuffer) Grow(n int) {
	if n < 0 {
		panic("jwriter: cannot grow buffer by a negative count")
	}
	b.reserve(n)
}

// reserve ensures that the buffer has room for at least n more bytes (n must not be
// negative), at least doubling the capacity when it must reallocate so that repeated
// appends remain amortized.
func (b *streamableBuffer) reserve(n int) {
	if cap(b.buf)-len(b.buf) < n {
		if len(b.buf)+n < 0 {
			panic("jwriter: buffer too large")
		}
		newCap := 2 * cap(b.buf)
		if newCap < len(b.buf)+n {
			newCap = len(b.buf) + n
		}
		newBuf := make([]byte, len(b.buf), newCap)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}

func (b *streamableBuffer) SetStreamingWriter(w io.Writer, chunkSize int) {
	b.dest = w
	b.chunkSize = chunkSize
}

// Flush writes any buffered output to the destination, if there is one. Once a destination
// write has failed, the failure is remembered and returned by every subsequent Flush call,
// even after the buffered data that could not be delivered has been discarded.
func (b *streamableBuffer) Flush() error {
	if b.dest == nil {
		return nil
	}
	if len(b.buf) > 0 {
		if b.destErr == nil {
			n, err := b.dest.Write(b.buf)
			if err == nil && n < len(b.buf) {
				err = io.ErrShortWrite
			}
			b.destErr = err
		}
		b.buf = b.buf[:0]
	}
	return b.destErr
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
	b.reserve(len(data))
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
	b.reserve(len(s))
	b.buf = append(b.buf, s...)
	b.maybeFlush()
}
