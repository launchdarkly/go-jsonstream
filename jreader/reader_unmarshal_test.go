package jreader

import (
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSONWithReader(t *testing.T) {
	var val ExampleStructWrapper
	err := UnmarshalJSONWithReader(commontest.ExampleStructData, &val)
	require.NoError(t, err)
	require.Equal(t, ExampleStructWrapper(commontest.ExampleStructValue), val)
}

func TestUnmarshalJSONWithReaderReturnsErrorForNonWhitespaceDatasAfterEnd(t *testing.T) {
	var val ExampleStructWrapper
	badJSON := string(commontest.ExampleStructData) + "xxx"
	err := UnmarshalJSONWithReader([]byte(badJSON), &val)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected data after end")
}

func TestUnmarshalJSONWithReaderDisregardsWhitespaceAfterEnd(t *testing.T) {
	var val ExampleStructWrapper
	okJSON := string(commontest.ExampleStructData) + "   \t\n\r  "
	err := UnmarshalJSONWithReader([]byte(okJSON), &val)
	require.NoError(t, err)
	require.Equal(t, ExampleStructWrapper(commontest.ExampleStructValue), val)
}
