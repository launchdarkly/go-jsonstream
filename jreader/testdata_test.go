package jreader

import (
	"github.com/launchdarkly/go-jsonstream/internal/commontest"
)

// ExampleStruct is defined in another package, so we need to wrap it in our own type to define methods on it.
type ExampleStructWrapper commontest.ExampleStruct

func (s *ExampleStructWrapper) ReadFromJSONReader(r *Reader) {
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case commontest.ExampleStructStringFieldName:
			s.StringField = r.String()
		case commontest.ExampleStructIntFieldName:
			s.IntField = r.Int()
		case commontest.ExampleStructOptBoolAsInterfaceFieldName:
			b, nonNull := r.BoolOrNull()
			if nonNull {
				s.OptBoolAsInterfaceField = b
			} else {
				s.OptBoolAsInterfaceField = nil
			}
		}
	}
}
