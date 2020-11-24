package jwriter

import (
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"
)

// This uses the framework defined in the commontest package to exercise Writer with a large
// number of JSON output permutations, in conjunction with DefaultTokenWriter.

type writerTestContext struct {
	w *Writer
}

type writerValueTestFactory struct{}

func TestWriter(t *testing.T) {
	s := commontest.WriterTestSuite{
		ContextFactory: func() commontest.TestContext {
			w := NewWriter()
			return &writerTestContext{
				w: &w,
			}
		},
		ValueTestFactory: writerValueTestFactory{},
		EncodeAsHex:      tokenWriterWillEncodeAsHex,
	}
	s.Run(t)
}

func (c writerTestContext) JSONData() []byte { return c.w.tw.Bytes() }

func (f writerValueTestFactory) EOF() commontest.Action {
	return func(c commontest.TestContext) error {
		return c.(*writerTestContext).w.Error()
	}
}

func (f writerValueTestFactory) Variants(value commontest.AnyValue) []commontest.ValueVariant {
	// Integer values can be written using either Int() or Float64().
	if value.Kind == commontest.NumberValue && float64(int(value.Number)) == value.Number {
		return variantsForWritingNumbers
	}
	return nil
}

func (f writerValueTestFactory) Value(value commontest.AnyValue, variant commontest.ValueVariant) commontest.Action {
	return func(c commontest.TestContext) error {
		ctx := c.(*writerTestContext)
		w := ctx.w

		switch value.Kind {
		case commontest.NullValue:
			w.Null()

		case commontest.BoolValue:
			w.Bool(value.Bool)

		case commontest.NumberValue:
			if variant == writeNumberAsInt {
				w.Int(int(value.Number))
			} else {
				w.Float64(value.Number)
			}

		case commontest.StringValue:
			w.String(value.String)

		case commontest.ArrayValue:
			arr := w.Array()
			for _, e := range value.Array {
				arr.Next()
				if err := e(ctx); err != nil {
					return err
				}
			}
			arr.End()

		case commontest.ObjectValue:
			obj := w.Object()
			for _, p := range value.Object {
				obj.Property(p.Name)
				if err := p.Action(ctx); err != nil {
					return err
				}
			}
			obj.End()
		}

		return w.Error()
	}
}
