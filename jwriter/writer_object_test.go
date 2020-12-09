package jwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectState(t *testing.T) {
	w := NewWriter()
	o := w.Object()

	o.Property("prop0")
	w.String("value0")

	o.Null("prop1")
	o.Bool("prop2", true)
	o.Int("prop3", 3)
	o.Float64("prop4", 4.5)
	o.String("prop5", "five")

	o.OptBool("no1", false, false)
	o.OptInt("no2", false, 9)
	o.OptFloat64("no3", false, 9.5)
	o.OptString("no4", false, "x")

	o.OptBool("prop6", true, false)
	o.OptInt("prop7", true, 9)
	o.OptFloat64("prop8", true, 9.5)
	o.OptString("prop9", true, "x")

	oa := o.Array("propa")
	oa.Int(10)
	oa.End()

	oo := o.Object("propo")
	oo.Int("eleven", 11)
	oo.End()

	o.End()

	require.NoError(t, w.Error())
	expected := `{"prop0":"value0","prop1":null,"prop2":true,"prop3":3,"prop4":4.5,"prop5":"five",` +
		`"prop6":false,"prop7":9,"prop8":9.5,"prop9":"x","propa":[10],"propo":{"eleven":11}}`
	assert.JSONEq(t, expected, string(w.Bytes()))
}
