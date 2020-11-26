package jreader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddErrorStopsObjectParsing(t *testing.T) {
	r := NewReader([]byte(`{"a":1, "b":2}`))
	obj := r.Object()
	require.True(t, obj.Next())
	require.Equal(t, "a", string(obj.Name()))
	require.Equal(t, 1, r.Int())

	err := errors.New("sorry")
	r.AddError(err)
	require.Equal(t, err, r.Error())

	require.False(t, obj.Next())
	require.Equal(t, err, r.Error())
}

func TestSyntaxErrorStopsObjectParsing(t *testing.T) {
	r := NewReader([]byte(`{"a":1, x "b":2}`))
	obj := r.Object()
	require.True(t, obj.Next())
	require.Equal(t, "a", string(obj.Name()))
	require.Equal(t, 1, r.Int())

	require.False(t, obj.Next())
	require.Error(t, r.Error())
}
