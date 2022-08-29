package jwriter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/internal/commontest"
)

// This uses the framework defined in WriterTestSuite to exercise any TokenWriter implementation
// with a large number of JSON output permutations.

type tokenWriterTestContext struct {
	tw        *tokenWriter
	getOutput func() []byte
}

type tokenWriterValueTestFactory struct{}

const (
	writeNumberAsInt commontest.ValueVariant = "int:"
)

var (
	variantsForScalarValues   = []commontest.ValueVariant{"", commontest.UntypedVariant}
	variantsForWritingNumbers = []commontest.ValueVariant{"", writeNumberAsInt, commontest.UntypedVariant}
)

type tokenWriterTestSuite struct {
	Factory     func() (*tokenWriter, func() []byte)
	EncodeAsHex func(rune) bool
}

func TestTokenWriter(t *testing.T) {
	s := tokenWriterTestSuite{
		Factory: func() (*tokenWriter, func() []byte) {
			tw := newTokenWriter()
			return &tw, tw.Bytes
		},
	}
	s.Run(t)
}

func (s tokenWriterTestSuite) Run(t *testing.T) {
	ws := commontest.WriterTestSuite{
		ContextFactory: func() commontest.TestContext {
			tw, getOutput := s.Factory()
			return &tokenWriterTestContext{
				tw:        tw,
				getOutput: getOutput,
			}
		},
		ValueTestFactory: tokenWriterValueTestFactory{},
		EncodeAsHex:      tokenWriterWillEncodeAsHex,
	}
	ws.Run(t)
}

func (c tokenWriterTestContext) JSONData() []byte { return c.getOutput() }

func (f tokenWriterValueTestFactory) EOF() commontest.Action {
	return func(c commontest.TestContext) error { return nil }
}

func (f tokenWriterValueTestFactory) Variants(value commontest.AnyValue) []commontest.ValueVariant {
	// Integer values can be written using either Int() or Float64().
	if value.Kind == commontest.NumberValue && float64(int(value.Number)) == value.Number {
		return variantsForWritingNumbers
	}
	if value.Kind != commontest.ArrayValue && value.Kind != commontest.ObjectValue {
		return variantsForScalarValues
	}
	return nil
}

func (f tokenWriterValueTestFactory) Value(value commontest.AnyValue, variant commontest.ValueVariant) commontest.Action {
	return func(c commontest.TestContext) error {
		ctx := c.(*tokenWriterTestContext)
		tw := ctx.tw

		switch value.Kind {
		case commontest.NullValue:
			if variant == commontest.UntypedVariant {
				return tw.Raw(json.RawMessage(`null`))
			}
			return tw.Null()

		case commontest.BoolValue:
			if variant == commontest.UntypedVariant {
				return tw.Raw(json.RawMessage(fmt.Sprintf("%t", value.Bool)))
			}
			return tw.Bool(value.Bool)

		case commontest.NumberValue:
			if variant == commontest.UntypedVariant {
				return tw.Raw(json.RawMessage(strconv.FormatFloat(value.Number, 'f', -1, 64)))
			}
			if variant == writeNumberAsInt {
				return tw.Int(int(value.Number))
			} else {
				return tw.Float64(value.Number)
			}

		case commontest.StringValue:
			if variant == commontest.UntypedVariant {
				// Use our own encoder to encode the string, but then write it with Raw()
				tw1 := newTokenWriter()
				_ = tw1.String(value.String)
				return tw.Raw(json.RawMessage(tw1.Bytes()))
			}
			return tw.String(value.String)

		case commontest.ArrayValue:
			if err := tw.Delimiter('['); err != nil {
				return err
			}
			first := true
			for _, e := range value.Array {
				if !first {
					if err := tw.Delimiter(','); err != nil {
						return err
					}
				}
				first = false
				if err := e(c); err != nil {
					return err
				}
			}
			return tw.Delimiter(']')

		case commontest.ObjectValue:
			if err := tw.Delimiter('{'); err != nil {
				return err
			}
			first := true
			for _, p := range value.Object {
				if !first {
					if err := tw.Delimiter(','); err != nil {
						return err
					}
				}
				first = false
				if err := tw.String(p.Name); err != nil {
					return err
				}
				if err := tw.Delimiter(':'); err != nil {
					return err
				}
				if err := p.Action(c); err != nil {
					return err
				}
			}
			return tw.Delimiter('}')
		}
		return nil
	}
}
