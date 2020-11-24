package jreader

// Reader is a high-level API for reading JSON data sequentially.
//
// It is designed to make writing custom unmarshallers for application types as convenient as
// possible. The general usage pattern is as follows:
//
// - Values are parsed in the order that they appear.
//
// - In general, the caller should know what data type is expected. Since it is common for
// properties to be nullable, the methods for reading scalar types have an option for allowing
// a null instead of the specified type. If the type is completely unknown, use Any.
//
// - For reading array or object structures, the Array and Object methods return a struct that
// keeps track of additional reader state while that structure is being parsed.
//
// - If any method encounters an error (due to either malformed JSON, or well-formed JSON that
// did not match the caller's data type expectations), the Reader permanently enters a failed
// state and remembers that error; all subsequent method calls will return the same error and no
// more parsing will happen. This means that the caller does not necessarily have to check the
// error return value of any individual method, although it can.
//
// The underlying low-level stream reading and JSON tokenizing logic is abstracted out with the
// TokenReader interface.
type Reader struct {
	tr                tokenReader
	awaitingReadValue bool // used by ArrayState & ObjectState
	err               error
}

// Error returns the first error that the Reader encountered, if the Reader is in a failed state,
// or nil if it is still in a good state.
func (r *Reader) Error() error {
	return r.err
}

// RequireEOF returns nil if all of the input has been consumed (not counting whitespace), or an
// error if not.
func (r *Reader) RequireEOF() error {
	if !r.tr.EOF() {
		return SyntaxError{Message: errMsgDataAfterEnd, Offset: r.tr.LastPos()}
	}
	return nil
}

// AddError sets the Reader's error value and puts it into a failed state. If the parameter is nil
// or the Reader was already in a failed state, it does nothing.
func (r *Reader) AddError(err error) {
	if r.err == nil {
		r.err = err
	}
}

// Null attempts to read a null value, returning an error if the next token is not a null.
func (r *Reader) Null() error {
	r.awaitingReadValue = false
	if r.err != nil {
		return r.err
	}
	isNull, err := r.tr.Null()
	if isNull || err != nil {
		return err
	}
	return r.typeErrorForCurrentToken(NullValue, false)
}

// Bool attempts to read a boolean value. If allowNull is true, it allows the value to be null
// instead; in the case of a null, the first and second return values are both false. If the value
// is not null, the second return value is true.
func (r *Reader) Bool(allowNull bool) (value bool, nonNull bool, err error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return false, false, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if isNull || err != nil {
			return false, false, err
		}
	}
	val, err := r.tr.Bool()
	if err != nil {
		err = adjustTypeError(err, allowNull)
		r.err = err
	}
	return val, true, err
}

// Int attempts to read a numeric value and returns it as an int. If allowNull is true, it allows
// the value to be null instead; in the case of a null, the first return value is zero and the
// second is false. If the value is not null, the second return value is true.
//
// Types other than number and null will cause an error; they are not converted to numbers.
func (r *Reader) Int(allowNull bool) (int, bool, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return 0, false, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if isNull || err != nil {
			return 0, false, err
		}
	}
	val, err := r.tr.Number()
	if err != nil {
		err = adjustTypeError(err, allowNull)
		r.err = err
	}
	return int(val), true, err
}

// Float64 attempts to read a numeric value and returns it as a float64. If allowNull is true, it
// allows the value to be null instead; in the case of a null, the first return value is zero and the
// second is false. If the value is not null, the second return value is true.
//
// Types other than number and null will cause an error; they are not converted to numbers.
func (r *Reader) Float64(allowNull bool) (float64, bool, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return 0, false, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if isNull || err != nil {
			return 0, false, err
		}
	}
	val, err := r.tr.Number()
	if err != nil {
		err = adjustTypeError(err, allowNull)
		r.err = err
	}
	return val, true, err
}

// String attempts to read a string value. If allowNull is true, it allows the value to be null instead;
// in the case of a null, the first return value is an empty string and the second is false. If the value
// is not null, the second return value is true.
//
// Types other than string and null will cause an error; they are not converted to strings.
func (r *Reader) String(allowNull bool) (string, bool, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return "", false, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if isNull || err != nil {
			return "", false, err
		}
	}
	val, err := r.tr.String()
	if err != nil {
		err = adjustTypeError(err, allowNull)
		r.err = err
	}
	return val, true, err

}

// Array attempts to begin reading an JSON array value. If allowNull is true, it allows the value to
// be null instead; in the case of a null, the ArrayState's IsDefined method will return false and it
// will behave as an empty array in all other ways. Types other than array and null will cause an error.
//
// The ArrayState is used only for advancing to the next item, and contains the necessary state for
// keeping track of this (such as expecting a comma expected before each item except the first). To
// read the value of each array element, you will still use the Reader's methods.
//
// See ArrayState for example code.
func (r *Reader) Array(allowNull bool) (ArrayState, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return ArrayState{}, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if err != nil {
			return ArrayState{}, err
		}
		if isNull {
			return ArrayState{}, nil
		}
	}
	gotDelim, err := r.tr.Delimiter('[')
	if err != nil {
		return ArrayState{}, err
	}
	if gotDelim {
		return ArrayState{r: r}, nil
	}
	return ArrayState{}, r.typeErrorForCurrentToken(ArrayValue, allowNull)
}

// Object attempts to begin reading an JSON object value. If allowNull is true, it allows the value to
// be null instead; in the case of a null, the ObjectState's IsDefined method will return true and it
// will behave as an empty object in all other ways. Types other than object and null will cause an error.
//
// The ObjectState is used only for advancing to the next item, and contains the necessary state for
// keeping track of this (such as expecting a comma before each item except the first, and keeping
// track of the current property name). To read the value of each property, you will still use the
// Reader's methods.
//
// See ObjectState for example code.
func (r *Reader) Object(allowNull bool) (ObjectState, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return ObjectState{}, r.err
	}
	if allowNull {
		isNull, err := r.tr.Null()
		if err != nil {
			return ObjectState{}, err
		}
		if isNull {
			return ObjectState{}, nil
		}
	}
	gotDelim, err := r.tr.Delimiter('{')
	if err != nil {
		return ObjectState{}, err
	}
	if gotDelim {
		return ObjectState{r: r}, nil
	}
	return ObjectState{}, r.typeErrorForCurrentToken(ObjectValue, allowNull)
}

// Any reads a single value of any type, if it is a scalar value or a null, or prepares to read
// the value if it is an array or object.
//
// The returned AnyValue's Kind field indicates the value type. If it is BoolValue, NumberValue,
// or StringValue, check the corresponding Bool, Number, or String property. If it is ArrayValue
// or ObjectValue, the AnyValue's Array or Object field has been initialized with an ArrayState or
// ObjectState just as if you had called the Reader's Array or Object method.
func (r *Reader) Any() (AnyValue, error) {
	r.awaitingReadValue = false
	if r.err != nil {
		return AnyValue{}, r.err
	}
	v, err := r.tr.Any()
	if err != nil {
		r.err = err
		return AnyValue{}, err
	}
	switch v.Kind {
	case BoolValue:
		return AnyValue{Kind: v.Kind, Bool: v.Bool}, nil
	case NumberValue:
		return AnyValue{Kind: v.Kind, Number: v.Number}, nil
	case StringValue:
		return AnyValue{Kind: v.Kind, String: v.String}, nil
	case ArrayValue:
		return AnyValue{Kind: v.Kind, Array: ArrayState{r: r}}, nil
	case ObjectValue:
		return AnyValue{Kind: v.Kind, Object: ObjectState{r: r}}, nil
	default:
		return AnyValue{Kind: NullValue}, nil
	}
}

// SkipValue consumes and discards the next JSON value of any type. For an array or object value, it
// recurses to also consume and discard all array elements or object properties.
func (r *Reader) SkipValue() error {
	r.awaitingReadValue = false
	if r.err != nil {
		return r.err
	}
	v, err := r.Any()
	if err != nil {
		r.err = err
		return err
	}
	if v.Kind == ArrayValue {
		for v.Array.Next() {
		}
	} else if v.Kind == ObjectValue {
		for v.Object.Next() {
		}
	}
	return r.err
}

func adjustTypeError(err error, nullable bool) error {
	if err != nil {
		switch e := err.(type) {
		case TypeError:
			e.Nullable = nullable
			return e
		}
	}
	return err
}

func (r *Reader) typeErrorForCurrentToken(expected ValueKind, nullable bool) error {
	v, err := r.tr.Any()
	if err != nil {
		return nil
	}
	return TypeError{Expected: expected, Actual: v.Kind, Offset: r.tr.LastPos(), Nullable: nullable}
}
