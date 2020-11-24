package jwriter

// ArrayWriter is a decorator that writes values to an underlying Writer within the context of a
// JSON array, adding commas between values as appropriate.
type ArrayState struct {
	w        *Writer
	hasItems bool
}

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

func (arr *ArrayState) End() {
	if arr.w == nil || arr.w.err != nil {
		return
	}
	arr.w.AddError(arr.w.tw.Delimiter(']'))
	arr.w = nil
}

// func (aw *ArrayWriter) beforeValue() bool {
// 	if aw.w == nil || aw.w.err != nil {
// 		return false
// 	}
// 	if aw.hasItems {
// 		if err := aw.w.tw.Delimiter(','); err != nil {
// 			aw.w.AddError(err)
// 			return false
// 		}
// 	}
// 	aw.hasItems = true
// 	return true
// }

// func (aw *ArrayWriter) End() {
// 	if aw.w != nil {
// 		aw.w.AddError(aw.w.tw.Delimiter(']'))
// 		aw.w = nil
// 	}
// }

// func (aw *ArrayWriter) Error() error {
// 	if aw.w == nil {
// 		return nil
// 	}
// 	return aw.w.Error()
// }

// func (aw *ArrayWriter) AddError(err error) {
// 	if aw.w != nil {
// 		aw.w.AddError(err)
// 	}
// }

// func (aw *ArrayWriter) Null() {
// 	if aw.beforeValue() {
// 		aw.w.Null()
// 	}
// }

// func (aw *ArrayWriter) Bool(value bool) {
// 	if aw.beforeValue() {
// 		aw.w.Bool(value)
// 	}
// }

// func (aw *ArrayWriter) Int(value int) {
// 	if aw.beforeValue() {
// 		aw.w.Int(value)
// 	}
// }

// func (aw *ArrayWriter) Float64(value float64) {
// 	if aw.beforeValue() {
// 		aw.w.Float64(value)
// 	}
// }

// func (aw *ArrayWriter) String(value string) {
// 	if aw.beforeValue() {
// 		aw.w.String(value)
// 	}
// }

// func (aw *ArrayWriter) Raw(value json.RawMessage) {
// 	if aw.beforeValue() {
// 		aw.w.Raw(value)
// 	}
// }

// func (aw *ArrayWriter) Array() ArrayWriter {
// 	if aw.beforeValue() {
// 		return aw.w.Array()
// 	}
// 	return ArrayWriter{}
// }

// func (aw *ArrayWriter) Object() ObjectWriter {
// 	if aw.beforeValue() {
// 		return aw.w.Object()
// 	}
// 	return ObjectWriter{}
// }
