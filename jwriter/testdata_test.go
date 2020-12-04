package jwriter

import (
	"github.com/launchdarkly/go-jsonstream/internal/commontest"
)

// ExampleStruct is defined in another package, so we need to wrap it in our own type to define methods on it.
type ExampleStructWrapper commontest.ExampleStruct

func (s ExampleStructWrapper) WriteToJSONWriter(w *Writer) {
	obj := w.Object()
	obj.String(commontest.ExampleStructStringFieldName, s.StringField)
	obj.Int(commontest.ExampleStructIntFieldName, s.IntField)
	obj.Property(commontest.ExampleStructOptBoolAsInterfaceFieldName)
	if s.OptBoolAsInterfaceField == nil {
		w.Null()
	} else {
		w.Bool(s.OptBoolAsInterfaceField.(bool))
	}
	obj.End()
}
