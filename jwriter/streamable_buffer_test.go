package jwriter

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamableBufferInMemoryMode(t *testing.T) {
	var b streamableBuffer
	expected := writeTestDataToBuffer(&b)
	assert.Equal(t, expected, string(b.Bytes()))
}

func TestStreamableBufferFlushDoesNothingByDefault(t *testing.T) {
	var b streamableBuffer
	expected := writeTestDataToBuffer(&b)
	require.NoError(t, b.Flush())
	assert.Equal(t, expected, string(b.Bytes()))
}

func TestStreamableBufferStreamingMode(t *testing.T) {
	t.Run("verify full data", func(t *testing.T) {
		var b streamableBuffer
		var target bytes.Buffer
		b.SetStreamingWriter(&target, 20)
		expected := writeTestDataToBuffer(&b)
		b.Flush()
		assert.Equal(t, expected, target.String())
	})

	t.Run("data is flushed incrementally", func(t *testing.T) {
		var b streamableBuffer
		var target bytes.Buffer
		b.SetStreamingWriter(&target, 10)

		b.WriteString("12345678")
		assert.Len(t, target.Bytes(), 0)

		b.WriteString("90")
		assert.Equal(t, "1234567890", target.String())

		b.WriteString("abcdefghijklm")
		assert.Equal(t, "1234567890abcdefghijklm", target.String())

		b.WriteString("nopqrstu")
		assert.Equal(t, "1234567890abcdefghijklm", target.String())

		b.WriteRune('v')
		b.WriteByte('w')
		assert.Equal(t, "1234567890abcdefghijklmnopqrstuvw", target.String())

		b.WriteString("xyz")
		b.Flush()
		assert.Equal(t, "1234567890abcdefghijklmnopqrstuvwxyz", target.String())
	})
}

func writeTestDataToBuffer(b *streamableBuffer) string {
	s := "abcdefghijklmnopqrstuvwxyz🐈"
	expected := ""

	for i := 0; i < 100; i++ {
		b.WriteString(s)
		expected += s
		b.WriteRune('$')
		b.WriteByte(' ')
		expected += "$ "
	}

	return expected
}
