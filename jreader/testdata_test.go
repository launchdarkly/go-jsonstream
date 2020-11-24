package jreader

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"
)

var (
	benchmarkBoolValue     = true
	benchmarkBoolPointer   = &benchmarkBoolValue
	benchmarkIntValue      = 3333
	benchmarkIntPointer    = &benchmarkIntValue
	benchmarkStringValue   = "value"
	benchmarkStringPointer = &benchmarkStringValue

	benchmarkErrResult     error
	benchmarkStringResult  string
	benchmarkBoolResult    bool
	benchmarkIntResult     int
	benchmarkFloat64Result float64
	benchmarkJSONResult    []byte
)

// ExampleStruct is defined in another package, so we need to wrap it in our own type to define methods on it.
type ExampleStructWrapper commontest.ExampleStruct

func (s *ExampleStructWrapper) ReadFromJSONReader(r *Reader) error {
	obj, err := r.Object(false)
	if err != nil {
		return err
	}
	for obj.Next() {
		var err error
		switch string(obj.Name()) {
		case commontest.ExampleStructStringFieldName:
			s.StringField, _, err = r.String(false)
		case commontest.ExampleStructIntFieldName:
			s.IntField, _, err = r.Int(false)
		case commontest.ExampleStructOptBoolAsInterfaceFieldName:
			b, nonNull, err := r.Bool(true)
			if err != nil {
				return err
			}
			if nonNull {
				s.OptBoolAsInterfaceField = b
			} else {
				s.OptBoolAsInterfaceField = nil
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
