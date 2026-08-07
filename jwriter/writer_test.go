package jwriter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/go-jsonstream/v4/internal/commontest"
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
	if value.Kind != commontest.ArrayValue && value.Kind != commontest.ObjectValue {
		return variantsForScalarValues
	}
	return nil
}

func (f writerValueTestFactory) Value(value commontest.AnyValue, variant commontest.ValueVariant) commontest.Action {
	return func(c commontest.TestContext) error {
		ctx := c.(*writerTestContext)
		w := ctx.w

		switch value.Kind {
		case commontest.NullValue:
			if variant == commontest.UntypedVariant {
				w.Raw(json.RawMessage(`null`))
				return w.Error()
			}
			w.Null()

		case commontest.BoolValue:
			if variant == commontest.UntypedVariant {
				w.Raw(json.RawMessage(fmt.Sprintf("%t", value.Bool)))
				return w.Error()
			}
			w.Bool(value.Bool)

		case commontest.NumberValue:
			if variant == commontest.UntypedVariant {
				w.Raw(json.RawMessage(strconv.FormatFloat(value.Number, 'f', -1, 64)))
				return w.Error()
			}
			if variant == writeNumberAsInt {
				w.Int(int(value.Number))
			} else {
				w.Float64(value.Number)
			}

		case commontest.StringValue:
			if variant == commontest.UntypedVariant {
				// Use our own encoder to encode the string, but then write it with Raw()
				tw1 := newTokenWriter()
				_ = tw1.String(value.String)
				w.Raw(json.RawMessage(tw1.Bytes()))
				return w.Error()
			}
			w.String(value.String)

		case commontest.ArrayValue:
			arr := w.Array()
			for _, e := range value.Array {
				if err := e(ctx); err != nil {
					return err
				}
			}
			arr.End()

		case commontest.ObjectValue:
			obj := w.Object()
			for _, p := range value.Object {
				obj.Name(p.Name)
				if err := p.Action(ctx); err != nil {
					return err
				}
			}
			obj.End()
		}

		return w.Error()
	}
}

func TestWriterReallocationsAreAmortized(t *testing.T) {
	// Growing the output buffer must at least double its capacity when tokens are
	// written, so the number of reallocations stays logarithmic in the output size.
	// 5000 tokens of each kind produce far under 64 KiB of output, so a doubling ladder
	// starting at 64 bytes can take at most ~11 steps; the bound below fails if a token
	// path falls back to append's ~1.25x growth. (The bound has slack because delimiter
	// writes do not reserve; see the streamableBuffer comment.)
	const maxSteps = 14
	writeToken := map[string]func(arr *ArrayState, i int){
		"string": func(arr *ArrayState, i int) { arr.String("value\twith\n\"escaped chars\"") },
		"int":    func(arr *ArrayState, i int) { arr.Int(123456789 + i) },
		"bool":   func(arr *ArrayState, i int) { arr.Bool(i%2 == 0) },
		"raw":    func(arr *ArrayState, i int) { arr.Raw(json.RawMessage(`{"k":[1,2,3]}`)) },
	}
	for kind, write := range writeToken {
		t.Run(kind, func(t *testing.T) {
			w := NewWriter()
			arr := w.Array()
			steps := 0
			lastCap := -1
			for i := 0; i < 5000; i++ {
				write(&arr, i)
				if c := cap(w.tw.buf.Bytes()); c != lastCap {
					steps++
					lastCap = c
				}
			}
			arr.End()
			require.NoError(t, w.Error())
			require.LessOrEqual(t, steps, maxSteps)
		})
	}
}
