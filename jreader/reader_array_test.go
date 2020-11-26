package jreader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddErrorStopsArrayParsing(t *testing.T) {
	r := NewReader([]byte("[1,2]"))
	arr := r.Array()
	require.True(t, arr.Next())
	require.Equal(t, 1, r.Int())

	err := errors.New("sorry")
	r.AddError(err)
	require.Equal(t, err, r.Error())

	require.False(t, arr.Next())
	require.Equal(t, 0, r.Int())
	require.Equal(t, err, r.Error())
}

func TestSyntaxErrorStopsArrayParsing(t *testing.T) {
	r := NewReader([]byte("[1,x,2]"))
	arr := r.Array()
	require.True(t, arr.Next())
	require.Equal(t, 1, r.Int())

	require.False(t, arr.Next())
	require.Equal(t, 0, r.Int())
	require.Error(t, r.Error())
}
