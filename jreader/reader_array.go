package jreader

// ArrayState is returned by Reader's Array method. Use it in conjunction with Reader to iterate
// through a JSON array. To read the value of each array element, you will still use the Reader's
// methods.
//
// This example reads an array of strings. Note that it is not necessary to check the error value
// from Array, or to break out of the loop if String fails, because the ArrayState's Next method
// will return false if the Reader has had any errors.
//
//     var values []string
//     arr, _ := r.Array(false)
//     for arr.Next() {
//         if s, _, err := r.String(false); err == nil {
//             values = append(values, s)
//         }
//     }
type ArrayState struct {
	r          *Reader
	afterFirst bool
}

// IsDefined returns true if the ArrayState represents an actual array, or false if it was
// parsed from a null value or was the result of an error. If IsDefined is false, Next will
// always return false. The zero value ArrayState{} returns false for IsDefined.
func (arr *ArrayState) IsDefined() bool {
	return arr.r != nil
}

// Next checks whether an array element is available and returns true if so. It returns false
// if the Reader has reached the end of the array, or if any previous Reader operation failed,
// or if the array was empty or null.
//
// If Next returns true, you can then use Reader methods such as Bool or String to read the
// element value. If you do not care about the value, simply calling Next again without calling
// a Reader method will discard the value, just as if you had called SkipValue on the reader.
//
// See ArrayState for example code.
func (arr *ArrayState) Next() bool {
	if arr.r == nil || arr.r.err != nil {
		return false
	}
	var isEnd bool
	var err error
	if arr.afterFirst {
		if arr.r.awaitingReadValue {
			if err := arr.r.SkipValue(); err != nil {
				return false
			}
		}
		isEnd, err = arr.r.tr.EndDelimiterOrComma(']')
	} else {
		arr.afterFirst = true
		isEnd, err = arr.r.tr.Delimiter(']')
	}
	if err != nil {
		arr.r.AddError(err)
		return false
	}
	if !isEnd {
		arr.r.awaitingReadValue = true
	}
	return !isEnd
}
