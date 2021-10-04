//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package jwriter

// The expectations in this test are slightly different from the non-easyjson version in
// writer_streaming_default_test.go, because in the easyjson implementation we don't have quite as
// much control over when the buffer gets flushed, so it will only be flushed at the end of a write.

import (
	"bytes"
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
