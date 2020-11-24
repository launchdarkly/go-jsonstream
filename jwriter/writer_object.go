package jwriter

// ObjectWriter is a decorator that writes values to an underlying Writer within the context of a
// JSON object, adding property names and commas between values as appropriate.
type ObjectState struct {
	w        *Writer
	hasItems bool
}

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

func (obj *ObjectState) End() {
	if obj.w == nil || obj.w.err != nil {
		return
	}
	obj.w.AddError(obj.w.tw.Delimiter('}'))
	obj.w = nil
}

// func (ow *ObjectWriter) beforeValue() bool {
// 	if ow.w == nil || ow.w.err != nil || ow.propName == "" {
// 		return false
// 	}
// 	if ow.hasItems {
// 		if err := ow.w.tw.Delimiter(','); err != nil {
// 			ow.w.AddError(err)
// 			return false
// 		}
// 	}
// 	ow.hasItems = true
// 	ow.w.AddError(ow.w.tw.PropertyName(ow.propName))
// 	ow.propName = ""
// 	return true
// }

// func (ow *ObjectWriter) Property(name string) *ObjectWriter {
// 	ow.propName = name
// 	return ow
// }

// func (ow *ObjectWriter) End() {
// 	if ow.w != nil {
// 		ow.w.AddError(ow.w.tw.Delimiter('}'))
// 		ow.w = nil
// 	}
// }

// func (ow *ObjectWriter) Error() error {
// 	if ow.w == nil {
// 		return nil
// 	}
// 	return ow.w.Error()
// }

// func (ow *ObjectWriter) AddError(err error) {
// 	if ow.w != nil {
// 		ow.w.AddError(err)
// 	}
// }

// func (ow *ObjectWriter) Null() {
// 	if ow.beforeValue() {
// 		ow.w.Null()
// 	}
// }

// func (ow *ObjectWriter) Bool(value bool) {
// 	if ow.beforeValue() {
// 		ow.w.Bool(value)
// 	}
// }

// func (ow *ObjectWriter) Int(value int) {
// 	if ow.beforeValue() {
// 		ow.w.Int(value)
// 	}
// }

// func (ow *ObjectWriter) Float64(value float64) {
// 	if ow.beforeValue() {
// 		ow.w.Float64(value)
// 	}
// }

// func (ow *ObjectWriter) String(value string) {
// 	if ow.beforeValue() {
// 		ow.w.String(value)
// 	}
// }

// func (ow *ObjectWriter) Raw(value json.RawMessage) {
// 	if ow.beforeValue() {
// 		ow.w.Raw(value)
// 	}
// }

// func (ow *ObjectWriter) Array() ArrayWriter {
// 	if ow.beforeValue() {
// 		return ow.w.Array()
// 	}
// 	return ArrayWriter{}
// }

// func (ow *ObjectWriter) Object() ObjectWriter {
// 	if ow.beforeValue() {
// 		return ow.w.Object()
// 	}
// 	return ObjectWriter{}
// }
