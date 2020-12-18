package jwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectState(t *testing.T) {
	w := NewWriter()
	o := w.Object()

	o.Name("prop1").Bool(true)
	o.Maybe("prop2", true).Bool(true)
	o.Maybe("shouldNotWriteThis", false).Bool(true)

	oa := o.Name("nestedArray").Array()
	oa.Int(1)
	oa.End()

	oo := o.Name("nestedObject").Object()
	oo.Name("eleven").Int(11)
	oo.End()

	o.End()

	require.NoError(t, w.Error())
	expected := `{"prop1":true,"prop2":true,"nestedArray":[1],"nestedObject":{"eleven":11}}`
	assert.JSONEq(t, expected, string(w.Bytes()))
}
