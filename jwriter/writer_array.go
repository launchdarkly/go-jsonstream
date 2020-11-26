package jwriter

// ArrayWriter is a decorator that writes values to an underlying Writer within the context of a
// JSON array, adding commas between values as appropriate.
type ArrayState struct {
	w        *Writer
	hasItems bool
}

// Next prepares to write the next array item. You can then use Writer methods to write the value.
func (arr *ArrayState) Next() {
	if arr.w == nil || arr.w.err != nil {
		return
	}
	if arr.hasItems {
		if err := arr.w.tw.Delimiter(','); err != nil {
			arr.w.AddError(err)
		}
	} else {
		arr.hasItems = true
	}
}

// Bool is a shortcut for calling Next() followed by writer.Null().
func (arr *ArrayState) Null() {
	if arr.w != nil {
		arr.Next()
		arr.w.Null()
	}
}

// Bool is a shortcut for calling Next() followed by writer.Bool(value).
func (arr *ArrayState) Bool(value bool) {
	if arr.w != nil {
		arr.Next()
		arr.w.Bool(value)
	}
}

// Int is a shortcut for calling Next() followed by writer.Int(value).
func (arr *ArrayState) Int(value int) {
	if arr.w != nil {
		arr.Next()
		arr.w.Int(value)
	}
}

// Float64 is a shortcut for calling Next() followed by writer.Float64(value).
func (arr *ArrayState) Float64(value float64) {
	if arr.w != nil {
		arr.Next()
		arr.w.Float64(value)
	}
}

// String is a shortcut for calling Next() followed by writer.String(value).
func (arr *ArrayState) String(value string) {
	if arr.w != nil {
		arr.Next()
		arr.w.String(value)
	}
}

// Array is a shortcut for calling Next() followed by writer.Array(), to create a nested array.
func (arr *ArrayState) Array() ArrayState {
	if arr.w != nil {
		arr.Next()
		return arr.w.Array()
	}
	return ArrayState{}
}

// Object is a shortcut for calling Next() followed by writer.Object(), to create a nested object.
func (arr *ArrayState) Object() ObjectState {
	if arr.w != nil {
		arr.Next()
		return arr.w.Object()
	}
	return ObjectState{}
}

// End writes the closing delimiter of the array.
func (arr *ArrayState) End() {
	if arr.w == nil || arr.w.err != nil {
		return
	}
	arr.w.AddError(arr.w.tw.Delimiter(']'))
	arr.w = nil
}
