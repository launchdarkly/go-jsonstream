package jwriter

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"
)

var (
	benchmarkBytesResult []byte
	benchmarkErrResult   error
)

// ExampleStruct is defined in another package, so we need to wrap it in our own type to define methods on it.
type ExampleStructWrapper commontest.ExampleStruct

func (s ExampleStructWrapper) WriteToJSONWriter(w *Writer) {
	obj := w.Object()
	obj.Property(commontest.ExampleStructStringFieldName)
	w.String(s.StringField)
	obj.Property(commontest.ExampleStructIntFieldName)
	w.Int(s.IntField)
	obj.Property(commontest.ExampleStructOptBoolAsInterfaceFieldName)
	if s.OptBoolAsInterfaceField == nil {
		w.Null()
	} else {
		w.Bool(s.OptBoolAsInterfaceField.(bool))
	}
	obj.End()
}
