package jreader

import (
	"errors"
	"fmt"
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"

	"github.com/stretchr/testify/require"
)

// This uses the framework defined in the commontest package to exercise Reader with a large number
// of valid and invalid JSON inputs. All we need to implement here is the logic for calling the
// appropriate Reader methods that correspond to the commontest abstractions.

type readerTestContext struct {
	input []byte
	r     *Reader
}

type readerValueTestFactory struct{}
type readerErrorTestFactory struct{}

// The behavior of Reader is flexible so that callers can choose to read the same JSON value in
// several different ways. Therefore, we generate variants for each value test as follows:
// - A null JSON value could be read either as a null, or as a nullable value of another type.
// - A JSON number could be read as an int (if the test value is an int), a float, or a nullable
// int or float.
// - Any other non-null value could be read as its own type, or as a nullable value of that type.
// - Any value could be read with a nonspecific type using the Any() method.
// - Any value could be skipped instead of read.
const (
	defaultVariant       commontest.ValueVariant = ""
	nullableValue        commontest.ValueVariant = "nullable"
	numberAsInt          commontest.ValueVariant = "int"
	nullableNumberAsInt  commontest.ValueVariant = "nullable int"
	nullableBoolIsNull   commontest.ValueVariant = "nullable bool is"
	nullableIntIsNull    commontest.ValueVariant = "nullable int is"
	nullableFloatIsNull  commontest.ValueVariant = "nullable float is"
	nullableStringIsNull commontest.ValueVariant = "nullable string is"
	nullableArrayIsNull  commontest.ValueVariant = "nullable array is"
	nullableObjectIsNull commontest.ValueVariant = "nullable object is"
)

var variantsForNullValues = []commontest.ValueVariant{defaultVariant, nullableBoolIsNull, nullableIntIsNull,
	nullableFloatIsNull, nullableStringIsNull, nullableArrayIsNull, nullableObjectIsNull,
	commontest.ReadAsAnyTypeVariant, commontest.SkipValueVariant}
var variantsForInts = []commontest.ValueVariant{defaultVariant, numberAsInt, nullableValue, nullableNumberAsInt,
	commontest.ReadAsAnyTypeVariant, commontest.SkipValueVariant}
var variantsForFloats = []commontest.ValueVariant{defaultVariant, nullableValue,
	commontest.ReadAsAnyTypeVariant, commontest.SkipValueVariant}
var variantsForNonNullValues = []commontest.ValueVariant{defaultVariant, nullableValue,
	commontest.ReadAsAnyTypeVariant, commontest.SkipValueVariant}
var shouldHaveBeenNullError = errors.New("should have been null")
var shouldNotHaveBeenNullError = errors.New("should not have been null")

func TestReader(t *testing.T) {
	ts := commontest.ReaderTestSuite{
		ContextFactory: func(input []byte) commontest.TestContext {
			r := NewReader(input)
			return &readerTestContext{input: input, r: &r}
		},
		ValueTestFactory:     readerValueTestFactory{},
		ReadErrorTestFactory: readerErrorTestFactory{},
	}
	ts.Run(t)
}

func (c readerTestContext) JSONData() []byte { return c.input }

func (f readerValueTestFactory) EOF() commontest.Action {
	return func(c commontest.TestContext) error {
		return commontest.AssertTrue(c.(*readerTestContext).r.tr.EOF(), "unexpected data after end")
	}
}

func (f readerValueTestFactory) Variants(value commontest.AnyValue) []commontest.ValueVariant {
	switch value.Kind {
	case commontest.NullValue:
		return variantsForNullValues
	case commontest.NumberValue:
		if float64(int(value.Number)) == value.Number {
			return variantsForInts
		}
		return variantsForFloats
	default:
		return variantsForNonNullValues
	}
}

func (f readerValueTestFactory) Value(value commontest.AnyValue, variant commontest.ValueVariant) commontest.Action {
	return func(c commontest.TestContext) error {
		ctx := c.(*readerTestContext)
		r := ctx.r

		if variant == commontest.SkipValueVariant {
			return r.SkipValue()
		}
		if variant == commontest.ReadAsAnyTypeVariant {
			return assertReadAnyValue(ctx, r, value)
		}

		allowNull := variant == nullableValue || variant == nullableNumberAsInt

		switch value.Kind {
		case commontest.NullValue:
			return assertReadNull(r, variant)

		case commontest.BoolValue:
			gotVal, nonNull, err := r.Bool(allowNull)
			return commontest.AssertNoErrors(err,
				commontest.AssertTrue(nonNull, shouldNotHaveBeenNullError.Error()),
				commontest.AssertEqual(value.Bool, gotVal))

		case commontest.NumberValue:
			if variant == numberAsInt || variant == nullableNumberAsInt {
				gotVal, nonNull, err := r.Int(allowNull)
				return commontest.AssertNoErrors(err,
					commontest.AssertTrue(nonNull, shouldNotHaveBeenNullError.Error()),
					commontest.AssertEqual(int(value.Number), gotVal))
			} else {
				gotVal, nonNull, err := r.Float64(allowNull)
				return commontest.AssertNoErrors(err,
					commontest.AssertTrue(nonNull, shouldNotHaveBeenNullError.Error()),
					commontest.AssertEqual(value.Number, gotVal))
			}

		case commontest.StringValue:
			gotVal, nonNull, err := r.String(allowNull)
			return commontest.AssertNoErrors(err,
				commontest.AssertTrue(nonNull, shouldNotHaveBeenNullError.Error()),
				commontest.AssertEqual(value.String, gotVal))

		case commontest.ArrayValue:
			arr, err := r.Array(allowNull)
			if err != nil {
				return err
			}
			return assertReadArray(ctx, &arr, value)

		case commontest.ObjectValue:
			obj, err := r.Object(allowNull)
			if err != nil {
				return err
			}
			return assertReadObject(ctx, &obj, value)
		}
		return nil
	}
}

func assertReadNull(r *Reader, variant commontest.ValueVariant) error {
	var gotVal, expectVal interface{}
	var nonNull bool
	var err error
	switch variant {
	case defaultVariant:
		return r.Null()
	case nullableBoolIsNull:
		gotVal, nonNull, err = r.Bool(true)
		expectVal = false
	case nullableIntIsNull:
		gotVal, nonNull, err = r.Int(true)
		expectVal = 0
	case nullableFloatIsNull:
		gotVal, nonNull, err = r.Float64(true)
		expectVal = float64(0)
	case nullableStringIsNull:
		gotVal, nonNull, err = r.String(true)
		expectVal = ""
	case nullableArrayIsNull:
		arr, err := r.Array(true)
		if err != nil {
			return err
		}
		if arr.IsDefined() {
			return TypeError{Expected: NullValue, Actual: ArrayValue}
		}
		return nil
	case nullableObjectIsNull:
		obj, err := r.Object(true)
		if err != nil {
			return err
		}
		if obj.IsDefined() {
			return TypeError{Expected: NullValue, Actual: ObjectValue}
		}
		return nil
	}
	if err != nil {
		return err
	}
	return commontest.AssertNoErrors(commontest.AssertTrue(!nonNull, shouldHaveBeenNullError.Error()),
		commontest.AssertEqual(expectVal, gotVal))
}

func assertReadArray(ctx *readerTestContext, arr *ArrayState, value commontest.AnyValue) error {
	if err := commontest.AssertTrue(arr.IsDefined(), shouldNotHaveBeenNullError.Error()); err != nil {
		return err
	}
	for _, e := range value.Array {
		if err := commontest.AssertTrue(arr.Next(), "array ended too soon"); err != nil {
			return err
		}
		if err := e(ctx); err != nil {
			return err
		}
	}
	return commontest.AssertTrue(!arr.Next(), "expected end of array")
}

func assertReadObject(ctx *readerTestContext, obj *ObjectState, value commontest.AnyValue) error {
	if err := commontest.AssertTrue(obj.IsDefined(), "should not have been null"); err != nil {
		return err
	}
	for _, p := range value.Object {
		if err := commontest.AssertNoErrors(
			commontest.AssertTrue(obj.Next(), "object ended too soon"),
			commontest.AssertEqual(p.Name, string(obj.Name())),
		); err != nil {
			return err
		}
		if err := p.Action(ctx); err != nil {
			return err
		}
	}
	return commontest.AssertTrue(!obj.Next(), "expected end of object")
}

func assertReadAnyValue(ctx *readerTestContext, r *Reader, value commontest.AnyValue) error {
	av, err := r.Any()
	if err != nil {
		return err
	}

	switch value.Kind {
	case commontest.NullValue:
		return commontest.AssertEqual(NullValue, av.Kind)

	case commontest.BoolValue:
		return commontest.AssertNoErrors(commontest.AssertEqual(BoolValue, av.Kind),
			commontest.AssertEqual(value.Bool, av.Bool))

	case commontest.NumberValue:
		return commontest.AssertNoErrors(commontest.AssertEqual(NumberValue, av.Kind),
			commontest.AssertEqual(value.Number, av.Number))

	case commontest.StringValue:
		return commontest.AssertNoErrors(commontest.AssertEqual(StringValue, av.Kind),
			commontest.AssertEqual(value.String, av.String))

	case commontest.ArrayValue:
		if err := commontest.AssertEqual(ArrayValue, av.Kind); err != nil {
			return err
		}
		return assertReadArray(ctx, &av.Array, value)

	case commontest.ObjectValue:
		if err := commontest.AssertEqual(ObjectValue, av.Kind); err != nil {
			return err
		}
		return assertReadObject(ctx, &av.Object, value)
	}
	return nil
}

func (f readerErrorTestFactory) ExpectEOFError(err error) error {
	return tokenReaderErrorTestFactory{}.ExpectEOFError(err)
}

func (f readerErrorTestFactory) ExpectWrongTypeError(err error, expected commontest.ValueKind,
	variant commontest.ValueVariant, actual commontest.ValueKind) error {
	// Here our behavior is different from tokenReaderErrorTestFactory, because Reader has more possible
	// kinds of errors due to the convenience features for reading values as "some type *or* null".
	expectedError := TypeError{
		Expected: valueKindFromTestValueKind(expected),
		Actual:   valueKindFromTestValueKind(actual),
	}
	switch variant {
	case nullableValue, nullableNumberAsInt:
		expectedError.Nullable = true
	case nullableBoolIsNull:
		expectedError.Nullable = true
		expectedError.Expected = BoolValue
	case nullableIntIsNull, nullableFloatIsNull:
		expectedError.Nullable = true
		expectedError.Expected = NumberValue
	case nullableStringIsNull:
		expectedError.Nullable = true
		expectedError.Expected = StringValue
	case nullableArrayIsNull:
		expectedError.Nullable = true
		expectedError.Expected = ArrayValue
	case nullableObjectIsNull:
		expectedError.Nullable = true
		expectedError.Expected = ObjectValue
	}
	if expectedError.Nullable && valueKindFromTestValueKind(actual) == expectedError.Expected {
		return commontest.AssertEqual(shouldHaveBeenNullError, err)
	}
	if expectedError.Nullable && actual == commontest.NullValue {
		return commontest.AssertEqual(shouldNotHaveBeenNullError, err)
	}
	if te, ok := err.(TypeError); ok {
		expectedError.Offset = te.Offset
		if te == expectedError {
			return nil
		}
	}
	return fmt.Errorf("expected %T %+v, got %T %+v", expectedError, expectedError, err, err)
}

func (f readerErrorTestFactory) ExpectSyntaxError(err error) error {
	return tokenReaderErrorTestFactory{}.ExpectSyntaxError(err)
}

func TestReaderSkipValue(t *testing.T) {
	t.Run("Next() skips array element if it was not read", func(t *testing.T) {
		data := []byte(`["a", ["b1", "b2"], "c"]`)
		r := NewReader(data)
		arr, err := r.Array(false)
		require.NoError(t, err)

		require.True(t, arr.Next())
		val1, _, err := r.String(false)
		require.NoError(t, err)
		require.Equal(t, "a", val1)

		require.True(t, arr.Next())

		require.True(t, arr.Next())
		val3, _, err := r.String(false)
		require.NoError(t, err)
		require.Equal(t, "c", val3)

		require.False(t, arr.Next())
	})

	t.Run("Next() skips property value if it was not read", func(t *testing.T) {
		data := []byte(`{"a":1, "b":{"b1":2, "b2":3}, "c":4}`)
		r := NewReader(data)
		obj, err := r.Object(false)
		require.NoError(t, err)

		require.True(t, obj.Next())
		require.Equal(t, "a", string(obj.Name()))
		val1, _, err := r.Int(false)
		require.NoError(t, err)
		require.Equal(t, 1, val1)

		require.True(t, obj.Next())

		require.True(t, obj.Next())
		require.Equal(t, "c", string(obj.Name()))
		val3, _, err := r.Int(false)
		require.NoError(t, err)
		require.Equal(t, 4, val3)

		require.False(t, obj.Next())
	})
}
