package jwriter

import (
	"encoding/json"
	"io"
	"strconv"
)

// This file defines the low-level JSON token writer. We don't define an interface for these methods,
// because calling them through an interface would limit performance.

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

const hexDigits = "0123456789abcdef"

// initialBufferCapacity is the buffer capacity preallocated by newTokenWriter. Paying for one
// small allocation up front keeps the append growth ladder short for typical outputs; it is
// the same minimum that bytes.Buffer uses.
const initialBufferCapacity = 64

// maxNumberLength is the longest text that Int or Float64 can produce for one value: a
// float64 in Go's shortest 'g' representation needs at most 24 characters (for example
// "-1.7976931348623157e+308"), and an int64 needs at most 20.
const maxNumberLength = 24

type tokenWriter struct {
	buf streamableBuffer
}

func newTokenWriter() tokenWriter {
	return newTokenWriterWithCapacity(initialBufferCapacity)
}

func newTokenWriterWithCapacity(capacity int) tokenWriter {
	tw := tokenWriter{}
	tw.buf.buf = make([]byte, 0, capacity)
	return tw
}

func newStreamingTokenWriter(dest io.Writer, bufferSize int) tokenWriter {
	tw := tokenWriter{}
	tw.buf.Grow(bufferSize)
	tw.buf.SetStreamingWriter(dest, bufferSize)
	return tw
}

// Bytes returns the full encoded byte slice.
//
// If the buffer is in a failed state from a previous invalid operation, Bytes() returns any data written
// so far.
func (tw *tokenWriter) Bytes() []byte {
	return tw.buf.Bytes()
}

// Flush writes any remaining in-memory output to the underlying Writer, if this is a streaming buffer
// created with newStreamingTokenWriter. It has no effect otherwise.
func (tw *tokenWriter) Flush() error {
	return tw.buf.Flush()
}

// Null writes a JSON null.
func (tw *tokenWriter) Null() error {
	tw.buf.Write(tokenNull)
	return tw.buf.GetWriterError()
}

// Bool writes a JSON boolean.
func (tw *tokenWriter) Bool(value bool) error {
	var out []byte
	if value {
		out = tokenTrue
	} else {
		out = tokenFalse
	}
	tw.buf.Write(out)
	return tw.buf.GetWriterError()
}

// Int writes an integer JSON number.
func (tw *tokenWriter) Int(value int) error {
	tw.buf.reserve(maxNumberLength)
	tw.buf.buf = strconv.AppendInt(tw.buf.buf, int64(value), 10)
	tw.buf.maybeFlush()
	return tw.buf.GetWriterError()
}

// Float64 writes a JSON number.
func (tw *tokenWriter) Float64(value float64) error {
	if value == 0 {
		tw.buf.WriteByte('0')
	} else {
		i := int(value)
		if float64(i) == value {
			return tw.Int(i)
		}
		tw.buf.reserve(maxNumberLength)
		tw.buf.buf = strconv.AppendFloat(tw.buf.buf, value, 'g', -1, 64)
		tw.buf.maybeFlush()
	}
	return tw.buf.GetWriterError()
}

// String writes a JSON string.
func (tw *tokenWriter) String(value string) error {
	return tw.writeQuotedString(value)
}

// Raw writes a preformatted chunk of JSON data.
func (tw *tokenWriter) Raw(value json.RawMessage) error {
	tw.buf.Write(value)
	return tw.buf.GetWriterError()
}

// PropertyName writes a JSON object property name followed by a colon.
func (tw *tokenWriter) PropertyName(name string) error {
	if err := tw.String(name); err != nil {
		return err
	}
	tw.buf.WriteByte(':')
	return tw.buf.GetWriterError()
}

// Delimiter writes a single character which must be a valid JSON delimiter ('{', ',', etc.).
func (tw *tokenWriter) Delimiter(delimiter byte) error {
	tw.buf.WriteByte(delimiter)
	return tw.buf.GetWriterError()
}

func (tw *tokenWriter) writeQuotedString(s string) error {
	// This is basically the same logic used internally by json.Marshal: scan for the next byte
	// that requires escaping, and copy the whole clean segment before it in one append. Bytes
	// outside the ASCII range are copied through verbatim without being decoded, since this
	// writer only ever escapes control characters, quotes, and backslashes.
	//
	// In-memory mode reserves len(s)+2 up front — the exact encoded length when nothing needs
	// escaping; escape expansion beyond that grows through append. Streaming mode must not
	// reserve by input length: it reserves nothing and instead
	// flushes at chunk boundaries as it scans, so that the buffer stays near the chunk size
	// no matter how long or escape-heavy the input string is. A clean segment is still
	// appended in one piece, so a segment longer than the chunk size overshoots it by that
	// segment's length, which the buffer permits.
	if tw.buf.dest == nil {
		tw.buf.reserve(len(s) + 2)
	}
	// dst is a local copy of the buffer's slice header; it must be stored back before any
	// buffer method runs (and reloaded after), or the flushed bytes would reappear.
	dst := tw.buf.buf
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		aByte := s[i]
		if aByte >= ' ' && aByte != '"' && aByte != '\\' {
			continue
		}
		dst = append(dst, s[start:i]...)
		dst = appendEscapedChar(dst, aByte)
		start = i + 1
		if tw.buf.dest != nil && len(dst) >= tw.buf.chunkSize {
			tw.buf.buf = dst
			tw.buf.maybeFlush()
			dst = tw.buf.buf
		}
	}
	dst = append(dst, s[start:]...)
	dst = append(dst, '"')
	tw.buf.buf = dst
	tw.buf.maybeFlush()
	return tw.buf.GetWriterError()
}

func appendEscapedChar(dst []byte, ch byte) []byte {
	switch ch {
	case '\b':
		return append(dst, '\\', 'b')
	case '\t':
		return append(dst, '\\', 't')
	case '\n':
		return append(dst, '\\', 'n')
	case '\f':
		return append(dst, '\\', 'f')
	case '\r':
		return append(dst, '\\', 'r')
	case '"':
		return append(dst, '\\', '"')
	case '\\':
		return append(dst, '\\', '\\')
	default:
		return append(dst, '\\', 'u', '0', '0', hexDigits[ch>>4], hexDigits[ch&0xf])
	}
}
