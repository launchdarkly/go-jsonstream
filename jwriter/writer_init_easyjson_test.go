// +build launchdarkly_easyjson

package jwriter

import (
	"testing"

	ejwriter "github.com/mailru/easyjson/jwriter"
	"github.com/stretchr/testify/require"
)

func TestNewWriterFromEasyjsonWriter(t *testing.T) {
	expectedOutput := `[1,{"property":2},3]`
	ejw := ejwriter.Writer{}

	// Write the first part of a JSON array directly with the easyjson Writer
	ejw.RawByte('[')
	require.NoError(t, ejw.Error)
	ejw.Int(1)
	require.NoError(t, ejw.Error)
	ejw.RawByte(',')
	require.NoError(t, ejw.Error)

	// Now pick up where we left off and use our Writer to write {"property":2}
	writer := NewWriterFromEasyjsonWriter(&ejw)
	obj := writer.Object()
	require.NoError(t, writer.Error())
	obj.Property("property")
	require.NoError(t, writer.Error())
	writer.Int(2)
	require.NoError(t, writer.Error())
	obj.End()

	// The easyjson Writer should be left in the proper state to write the rest of the stream
	require.NoError(t, ejw.Error)
	ejw.RawByte(',')
	require.NoError(t, ejw.Error)
	ejw.Int(3)
	require.NoError(t, ejw.Error)
	ejw.RawByte(']')
	require.NoError(t, ejw.Error)

	output, err := ejw.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, expectedOutput, string(output))
}
