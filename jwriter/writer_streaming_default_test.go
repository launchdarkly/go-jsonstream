package jwriter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamingWriterWritesToTargetInChunks(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := NewStreamingWriter(buf, 10)

	expected := ""

	arr := w.Array()
	require.Equal(t, expected, buf.String())

	arr.Bool(true)
	require.Equal(t, expected, buf.String())

	arr.String("abc")
	expected += `[true,"abc"`
	require.Equal(t, expected, buf.String())

	arr.Int(33)
	require.Equal(t, expected, buf.String())

	arr.Null()
	require.Equal(t, expected, buf.String())

	arr.Float64(2.5)
	expected += `,33,null,2.5`
	require.Equal(t, expected, buf.String())

	arr.End()
	require.Equal(t, expected, buf.String())

	require.NoError(t, w.Flush())
	expected += `]`
	require.Equal(t, expected, buf.String())
}

type failingDestination struct {
	failAfter int // number of successful writes before failures begin
	writes    int
}

func (f *failingDestination) Write(p []byte) (int, error) {
	f.writes++
	if f.writes > f.failAfter {
		return 0, errors.New("destination failed")
	}
	return len(p), nil
}

func writeStreamingTestDocument(w *Writer) {
	arr := w.Array()
	arr.String("hello world")
	arr.Int(1)
	arr.String("value\twith\n\"escaped chars\"")
	arr.Float64(2.5)
	arr.Null()
	arr.Bool(true)
	obj := arr.Object()
	obj.Name("prop").Int(-1234567890)
	obj.End()
	arr.Raw(json.RawMessage(`{"raw":[1,2,3]}`))
	arr.End()
}

func TestStreamingWriterSurfacesDestinationError(t *testing.T) {
	// Whenever the destination has failed, both Error() and the final Flush() must report it,
	// including when a failed mid-stream flush left the buffer empty.
	cases := []struct{ chunkSize, failAfter int }{
		{1, 0}, {1, 1}, {1, 2}, {1, 5},
		{2, 0}, {2, 1},
		{10, 0}, {10, 1},
		{50, 0},
		{1000, 0}, // nothing flushes until the final Flush
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("chunkSize=%d failAfter=%d", c.chunkSize, c.failAfter), func(t *testing.T) {
			f := &failingDestination{failAfter: c.failAfter}
			w := NewStreamingWriter(f, c.chunkSize)
			writeStreamingTestDocument(&w)
			require.Error(t, w.Flush())
			require.Error(t, w.Error())
		})
	}
}

func TestStreamingWriterOutputMatchesInMemoryWriter(t *testing.T) {
	expectedWriter := NewWriter()
	writeStreamingTestDocument(&expectedWriter)
	require.NoError(t, expectedWriter.Error())
	expected := string(expectedWriter.Bytes())

	for _, chunkSize := range []int{0, 1, 2, 3, 5, 8, 13, 64, 100, 1000} {
		t.Run(fmt.Sprintf("chunkSize=%d", chunkSize), func(t *testing.T) {
			var target bytes.Buffer
			w := NewStreamingWriter(&target, chunkSize)
			writeStreamingTestDocument(&w)
			require.NoError(t, w.Flush())
			require.Equal(t, expected, target.String())
		})
	}
}

func TestStreamingWriterFlushesAfterEachToken(t *testing.T) {
	// With a chunk size of 1, every completed token must reach the destination immediately;
	// this pins the invariant that each token-writing code path checks for a flush.
	var target bytes.Buffer
	w := NewStreamingWriter(&target, 1)

	arr := w.Array()
	arr.Int(12345)
	require.Equal(t, `[12345`, target.String())
	arr.Float64(2.5)
	require.Equal(t, `[12345,2.5`, target.String())
	arr.String("ab\tc")
	require.Equal(t, `[12345,2.5,"ab\tc"`, target.String())
	arr.Bool(true)
	require.Equal(t, `[12345,2.5,"ab\tc",true`, target.String())
	arr.Raw(json.RawMessage(`{}`))
	require.Equal(t, `[12345,2.5,"ab\tc",true,{}`, target.String())
	arr.End()
	require.Equal(t, `[12345,2.5,"ab\tc",true,{}]`, target.String())

	require.NoError(t, w.Flush())
	require.Equal(t, `[12345,2.5,"ab\tc",true,{}]`, target.String())
}

func TestStreamingWriterBoundsBufferForEscapeHeavyStrings(t *testing.T) {
	// A string full of escaped characters must not make the internal buffer grow in
	// proportion to the string: the writer flushes at chunk boundaries while scanning.
	const chunkSize = 1024
	const inputLen = 200000
	var target bytes.Buffer
	w := NewStreamingWriter(&target, chunkSize)
	w.String(strings.Repeat("\n", inputLen))
	require.NoError(t, w.Flush())
	require.Equal(t, `"`+strings.Repeat(`\n`, inputLen)+`"`, target.String())
	require.LessOrEqual(t, cap(w.tw.buf.Bytes()), 4*chunkSize)
}
