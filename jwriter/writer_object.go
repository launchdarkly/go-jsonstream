package jwriter

// ObjectWriter is a decorator that writes values to an underlying Writer within the context of a
// JSON object, adding property names and commas between values as appropriate.
type ObjectState struct {
	w        *Writer
	hasItems bool
}

// Property writes an object property name and a colon. You can then use Writer methods to write
// the property value.
func (obj *ObjectState) Property(name string) {
	if obj.w == nil || obj.w.err != nil {
		return
	}
	if obj.hasItems {
		if err := obj.w.tw.Delimiter(','); err != nil {
			obj.w.AddError(err)
			return
		}
	}
	obj.hasItems = true
	obj.w.AddError(obj.w.tw.PropertyName(name))
}

// Null is a shortcut for calling Property(name) followed by writer.Null().
func (obj *ObjectState) Null(name string) {
	if obj.w != nil {
		obj.Property(name)
		obj.w.Null()
	}
}

// Bool is a shortcut for calling Property(name) followed by writer.Bool(value).
func (obj *ObjectState) Bool(name string, value bool) {
	if obj.w != nil {
		obj.Property(name)
		obj.w.Bool(value)
	}
}

// Int is a shortcut for calling Property(name) followed by writer.Int(value).
func (obj *ObjectState) Int(name string, value int) {
	if obj.w != nil {
		obj.Property(name)
		obj.w.Int(value)
	}
}

// Float64 is a shortcut for calling Property(name) followed by writer.Float64(value).
func (obj *ObjectState) Float64(name string, value float64) {
	if obj.w != nil {
		obj.Property(name)
		obj.w.Float64(value)
	}
}

// String is a shortcut for calling Property(name) followed by writer.String(value).
func (obj *ObjectState) String(name string, value string) {
	if obj.w != nil {
		obj.Property(name)
		obj.w.String(value)
	}
}

// OptBool is a shortcut for calling Bool(name, value) if isDefined is true.
func (obj *ObjectState) OptBool(name string, isDefined bool, value bool) {
	if isDefined {
		obj.Bool(name, value)
	}
}

// OptInt is a shortcut for calling Int(name, value) if isDefined is true.
func (obj *ObjectState) OptInt(name string, isDefined bool, value int) {
	if isDefined {
		obj.Int(name, value)
	}
}

// OptFloat64 is a shortcut for calling Float64(name, value) if isDefined is true.
func (obj *ObjectState) OptFloat64(name string, isDefined bool, value float64) {
	if isDefined {
		obj.Float64(name, value)
	}
}

// OptString is a shortcut for calling String(name, value) if isDefined is true.
func (obj *ObjectState) OptString(name string, isDefined bool, value string) {
	if isDefined {
		obj.String(name, value)
	}
}

// Array is a shortcut for calling Property(name) followed by writer.Array(), to create a nested array.
func (obj *ObjectState) Array(name string) ArrayState {
	if obj.w != nil {
		obj.Property(name)
		return obj.w.Array()
	}
	return ArrayState{}
}

// Object is a shortcut for calling Property(name) followed by writer.Object(), to create a nested object.
func (obj *ObjectState) Object(name string) ObjectState {
	if obj.w != nil {
		obj.Property(name)
		return obj.w.Object()
	}
	return ObjectState{}
}

// End writes the closing delimiter of the object.
func (obj *ObjectState) End() {
	if obj.w == nil || obj.w.err != nil {
		return
	}
	obj.w.AddError(obj.w.tw.Delimiter('}'))
	obj.w = nil
}
