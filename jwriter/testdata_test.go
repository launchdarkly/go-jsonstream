package jwriter

import (
	"github.com/launchdarkly/go-jsonstream/v3/internal/commontest"
)

// ExampleStruct is defined in another package, so we need to wrap it in our own type to define methods on it.
type ExampleStructWrapper commontest.ExampleStruct

func (s ExampleStructWrapper) WriteToJSONWriter(w *Writer) {
	obj := w.Object()
	obj.Name(commontest.ExampleStructStringFieldName).String(s.StringField)
	obj.Name(commontest.ExampleStructIntFieldName).Int(s.IntField)
	obj.Name(commontest.ExampleStructOptBoolAsInterfaceFieldName)
	if s.OptBoolAsInterfaceField == nil {
		w.Null()
	} else {
		w.Bool(s.OptBoolAsInterfaceField.(bool))
	}
	obj.End()
}
