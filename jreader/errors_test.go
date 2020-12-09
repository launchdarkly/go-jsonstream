package jreader

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyntaxError(t *testing.T) {
	e1 := SyntaxError{Message: "xyz", Offset: 2}
	assert.Equal(t, "xyz at position 2", e1.Error())

	e2 := SyntaxError{Message: "xyz", Offset: 2, Value: "abc"}
	assert.Equal(t, `xyz at position 2 ("abc")`, e2.Error())
}

func TestTypeError(t *testing.T) {
	assert.Equal(t, "expected boolean, got string at position 2",
		TypeError{Expected: BoolValue, Actual: StringValue, Offset: 2}.Error())

	assert.Equal(t, "expected boolean or null, got string at position 2",
		TypeError{Expected: BoolValue, Actual: StringValue, Offset: 2, Nullable: true}.Error())

	assert.Equal(t, "expected null, got boolean at position 2",
		TypeError{Expected: NullValue, Actual: BoolValue, Offset: 2}.Error())

	assert.Equal(t, "expected null, got number at position 2",
		TypeError{Expected: NullValue, Actual: NumberValue, Offset: 2}.Error())

	assert.Equal(t, "expected null, got string at position 2",
		TypeError{Expected: NullValue, Actual: StringValue, Offset: 2}.Error())

	assert.Equal(t, "expected null, got array at position 2",
		TypeError{Expected: NullValue, Actual: ArrayValue, Offset: 2}.Error())

	assert.Equal(t, "expected null, got object at position 2",
		TypeError{Expected: NullValue, Actual: ObjectValue, Offset: 2}.Error())

	assert.Equal(t, "expected null, got unknown token at position 2",
		TypeError{Expected: NullValue, Actual: 99, Offset: 2}.Error())
}

func TestToJSONError(t *testing.T) {
	e1 := SyntaxError{Message: "xyz", Offset: 2}
	je1 := ToJSONError(e1, nil)
	assert.Equal(t, &json.SyntaxError{Offset: 2}, je1)

	e2 := TypeError{Expected: NumberValue, Actual: StringValue, Offset: 2}
	someIntValue := 1000
	je2 := ToJSONError(e2, someIntValue)
	assert.Equal(t, &json.UnmarshalTypeError{Value: "number", Offset: 2, Type: reflect.TypeOf(someIntValue)}, je2)

	e3 := errors.New("some other error")
	assert.Equal(t, e3, ToJSONError(e3, nil))
}
