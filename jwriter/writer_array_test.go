package jwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayState(t *testing.T) {
	w := NewWriter()
	a := w.Array()

	a.Null()
	a.Bool(true)
	a.Int(3)
	a.Float64(4.5)
	a.String("five")

	aa := a.Array()
	aa.Int(6)
	aa.End()

	ao := a.Object()
	ao.Name("seven").Int(7)
	ao.End()

	a.End()

	require.NoError(t, w.Error())
	expected := `[null,true,3,4.5,"five",[6],{"seven":7}]`
	assert.JSONEq(t, expected, string(w.Bytes()))
}
