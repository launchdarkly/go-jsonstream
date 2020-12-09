package jreader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
	r := NewReader([]byte(`{"a":1, x: 2, "c":3}`))
	obj := r.Object()
	require.True(t, obj.Next())
	require.Equal(t, "a", string(obj.Name()))
	require.Equal(t, 1, r.Int())

	require.False(t, obj.Next())
	require.Equal(t, 0, r.Int())

	require.Error(t, r.Error())

	require.False(t, obj.Next())
}

func TestRequiredPropertiesAreAllFound(t *testing.T) {
	data := []byte(`{"a":1, "b":2, "c":3}`)
	requiredProps := []string{"c", "b", "a"}
	r := NewReader(data)
	for obj := r.Object().WithRequiredProperties(requiredProps); obj.Next(); {
	}
	require.NoError(t, r.Error())
}

func TestRequiredPropertyIsNotFound(t *testing.T) {
	data := []byte(`{"a":1, "c":3}`)
	requiredProps := []string{"c", "b", "a"}
	r := NewReader(data)
	for obj := r.Object().WithRequiredProperties(requiredProps); obj.Next(); {
	}
	require.Error(t, r.Error())
	require.IsType(t, RequiredPropertyError{}, r.Error())
	rpe := r.Error().(RequiredPropertyError)
	assert.Equal(t, "b", rpe.Name)
	assert.GreaterOrEqual(t, rpe.Offset, len(data)-1)
}
