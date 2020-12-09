package jreader

// ObjectState is returned by Reader's Object and ObjectOrNull methods. Use it in conjunction with
// Reader to iterate through a JSON object. To read the value of each object property, you will
// still use the Reader's methods.
//
// This example reads an object whose values are strings; if there is a null instead of an object,
// it behaves the same as for an empty object. Note that it is not necessary to check for an error
// result before iterating over the ObjectState, or to break out of the loop if String causes an
// error, because the ObjectState's Next method will return false if the Reader has had any errors.
//
//     values := map[string]string
//     for obj := r.ObjectOrNull(); obj.Next(); {
//         key := string(obj.Name())
//         if s := r.String(); r.Error() == nil {
//             values[key] = s
//         }
//     }
//
// The next example reads an object with two expected property names, "a" and "b". Any unrecognized
// properties are ignored.
//
//     var result struct {
//         a int
//         b int
//     }
//     for obj := r.ObjectOrNull(); obj.Next(); {
//         switch string(obj.Name()) {
//         case "a":
//             result.a = r.Int()
//         case "b":
//             result.b = r.Int()
//         }
//     }
type ObjectState struct {
	r          *Reader
	afterFirst bool
	name       []byte
}

// IsDefined returns true if the ObjectState represents an actual object, or false if it was
// parsed from a null value or was the result of an error. If IsDefined is false, Next will
// always return false. The zero value ObjectState{} returns false for IsDefined.
func (obj *ObjectState) IsDefined() bool {
	return obj.r != nil
}

// Next checks whether an object property is available and returns true if so. It returns false
// if the Reader has reached the end of the object, or if any previous Reader operation failed,
// or if the object was empty or null.
//
// If Next returns true, you can then get the property name with Name, and use Reader methods
// such as Bool or String to read the property value. If you do not care about the value, simply
// calling Next again without calling a Reader method will discard the value, just as if you had
// called SkipValue on the reader.
//
// See ObjectState for example code.
func (obj *ObjectState) Next() bool {
	if obj.r == nil || obj.r.err != nil {
		return false
	}
	var isEnd bool
	var err error
	if obj.afterFirst {
		if obj.r.awaitingReadValue {
			if err := obj.r.SkipValue(); err != nil {
				return false
			}
		}
		isEnd, err = obj.r.tr.EndDelimiterOrComma('}')
	} else {
		obj.afterFirst = true
		isEnd, err = obj.r.tr.Delimiter('}')
	}
	if err != nil {
		obj.r.AddError(err)
		return false
	}
	if !isEnd {
		name, err := obj.r.tr.PropertyName()
		if err != nil {
			obj.r.AddError(err)
			return false
		}
		obj.name = name
		obj.r.awaitingReadValue = true
		return true
	}
	obj.name = nil
	return false
}

// Name returns the name of the current object property, or nil if there is no current property
// (that is, if Next returned false or if Next was never called).
//
// For efficiency, to avoid allocating a string for each property name, the name is returned as a
// byte slice which may refer directly to the source data. Casting this to a string within a simple
// comparison expression or switch statement should not cause a string allocation; the Go compiler
// optimizes these into direct byte-slice comparisons.
func (obj *ObjectState) Name() []byte {
	return obj.name
}
