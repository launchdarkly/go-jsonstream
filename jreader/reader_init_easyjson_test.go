//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package jreader

import (
	"testing"

	"github.com/mailru/easyjson/jlexer"
	"github.com/stretchr/testify/require"
)

func TestNewReaderFromEasyJSONLexer(t *testing.T) {
	data := []byte(`[1,{"property":2},3]`)
	lexer := jlexer.Lexer{Data: data}

	// Parse the first part of the JSON array directly with the Lexer
	lexer.Delim('[')
	require.NoError(t, lexer.Error())
	n := lexer.Int()
	require.Equal(t, 1, n)
	require.NoError(t, lexer.Error())
	lexer.WantComma()

	// Now pick up where we left off and use the Reader to parse {"property":2}
	reader := NewReaderFromEasyJSONLexer(&lexer)
	obj := reader.Object()
	require.NoError(t, reader.Error())
	require.True(t, obj.Next())
	require.Equal(t, "property", string(obj.Name()))
	n = reader.Int()
	require.NoError(t, reader.Error())
	require.Equal(t, 2, n)
	require.False(t, obj.Next())

	// The Lexer should be left in the proper state to parse the rest of the stream
	require.NoError(t, lexer.Error())
	lexer.WantComma()
	n = lexer.Int()
	require.Equal(t, 3, n)
	require.NoError(t, lexer.Error())
	lexer.WantComma() // we don't actually want a comma here, but that's how you read arrays in easyjson
	lexer.Delim(']')
	require.NoError(t, lexer.Error())
}
