//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jwriter

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
	expected += `[true,"abc`
	require.Equal(t, expected, buf.String())

	arr.Int(33)
	require.Equal(t, expected, buf.String())

	arr.Null()
	require.Equal(t, expected, buf.String())

	arr.Float64(2.5)
	expected += `",33,null,`
	require.Equal(t, expected, buf.String())

	arr.End()
	require.Equal(t, expected, buf.String())

	require.NoError(t, w.Flush())
	expected += `2.5]`
	require.Equal(t, expected, buf.String())
}
