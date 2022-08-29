package jreader

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/internal/commontest"
)

// This uses the framework defined in ReaderTestSuite to exercise any TokenReader implementation
// with a large number of valid and invalid JSON inputs.

type tokenReaderTestContext struct {
	input []byte
	tr    *tokenReader
}

type tokenReaderValueTestFactory struct{}
type tokenReaderErrorTestFactory struct{}

type TokenReaderTestSuite struct {
	Factory func([]byte) *tokenReader
}

func TestTokenReader(t *testing.T) {
	s := TokenReaderTestSuite{
		Factory: func(input []byte) *tokenReader {
			tr := newTokenReader(input)
			return &tr
		},
	}
	s.Run(t)
}

func (s TokenReaderTestSuite) Run(t *testing.T) {
	rs := commontest.ReaderTestSuite{
		ContextFactory: func(input []byte) commontest.TestContext {
			return &tokenReaderTestContext{
				input: input,
				tr:    s.Factory(input),
			}
		},
		ValueTestFactory:     tokenReaderValueTestFactory{},
		ReadErrorTestFactory: tokenReaderErrorTestFactory{},
	}
	rs.Run(t)
}

func (c tokenReaderTestContext) JSONData() []byte { return c.input }

func (f tokenReaderValueTestFactory) EOF() commontest.Action {
	return func(c commontest.TestContext) error {
		return commontest.AssertTrue(c.(*tokenReaderTestContext).tr.EOF(), "unexpected data after end")
	}
}

func (f tokenReaderValueTestFactory) Null() commontest.Action {
	return func(c commontest.TestContext) error {
		tr := c.(*tokenReaderTestContext).tr
		ok, err := tr.Null()
		if err != nil {
			return err
		}
		if !ok {
			any, _ := tr.Any()
			return TypeError{Expected: NullValue, Actual: any.Kind}
		}
		return nil
	}
}

func (f tokenReaderValueTestFactory) Variants(value commontest.AnyValue) []commontest.ValueVariant {
	// TokenReader does not need to use ValueVariants because it does not have any ambiguity
	// about what type a value is being read as.
	return nil
}

func (f tokenReaderValueTestFactory) Value(value commontest.AnyValue, variant commontest.ValueVariant) commontest.Action {
	return func(c commontest.TestContext) error {
		ctx := c.(*tokenReaderTestContext)
		tr := ctx.tr

		switch value.Kind {
		case commontest.NullValue:
			return f.Null()(c)

		case commontest.BoolValue:
			gotVal, err := tr.Bool()
			return commontest.AssertNoErrors(err, commontest.AssertEqual(value.Bool, gotVal))

		case commontest.NumberValue:
			gotVal, err := tr.Number()
			return commontest.AssertNoErrors(err, commontest.AssertEqual(value.Number, gotVal))

		case commontest.StringValue:
			gotVal, err := tr.String()
			return commontest.AssertNoErrors(err, commontest.AssertEqual(value.String, gotVal))

		case commontest.ArrayValue:
			gotDelim, err := tr.Delimiter('[')
			if err != nil {
				return err
			}
			if !gotDelim {
				return errors.New("expected start of array")
			}

			first := true
			for _, e := range value.Array {
				if !first {
					isEnd, err := tr.EndDelimiterOrComma(']')
					if err := commontest.AssertNoErrors(err, commontest.AssertTrue(!isEnd, "array ended too soon")); err != nil {
						return err
					}
				}
				first = false
				if err := e(c); err != nil {
					return err
				}
			}

			isEnd, err := tr.EndDelimiterOrComma(']')
			return commontest.AssertNoErrors(err, commontest.AssertTrue(isEnd, "expected end of array"))

		case commontest.ObjectValue:
			gotDelim, err := tr.Delimiter('{')
			if err != nil {
				return err
			}
			if !gotDelim {
				return errors.New("expected start of object")
			}

			first := true
			for _, p := range value.Object {
				if !first {
					isEnd, err := tr.EndDelimiterOrComma('}')
					if err := commontest.AssertNoErrors(err, commontest.AssertTrue(!isEnd, "object ended too soon")); err != nil {
						return err
					}
				}
				first = false
				name, err := tr.PropertyName()
				if err := commontest.AssertNoErrors(err, commontest.AssertEqual(string(name), p.Name)); err != nil {
					return err
				}
				if err := p.Action(c); err != nil {
					return err
				}
			}

			isEnd, err := tr.EndDelimiterOrComma('}')
			return commontest.AssertNoErrors(err, commontest.AssertTrue(isEnd, "expected end of object"))
		}
		return nil
	}
}

func (f tokenReaderErrorTestFactory) ExpectEOFError(err error) error {
	return commontest.AssertEqual(io.EOF, err)
}

func (f tokenReaderErrorTestFactory) ExpectWrongTypeError(err error, expected commontest.ValueKind,
	variant commontest.ValueVariant, actual commontest.ValueKind) error {
	if te, ok := err.(TypeError); ok {
		if te.Actual == valueKindFromTestValueKind(actual) && te.Expected == valueKindFromTestValueKind(expected) {
			return nil
		}
	}
	return fmt.Errorf("expected TypeError{Expected: %s, Actual: %s}, got %T %+v",
		valueKindFromTestValueKind(expected), valueKindFromTestValueKind(actual), err, err)
}

func (f tokenReaderErrorTestFactory) ExpectSyntaxError(err error) error {
	if _, ok := err.(SyntaxError); ok {
		return nil
	}
	return fmt.Errorf("expected SyntaxError, got %T %+v", err, err)
}

func valueKindFromTestValueKind(kind commontest.ValueKind) ValueKind {
	switch kind {
	case commontest.NullValue:
		return NullValue
	case commontest.BoolValue:
		return BoolValue
	case commontest.NumberValue:
		return NumberValue
	case commontest.StringValue:
		return StringValue
	case commontest.ArrayValue:
		return ArrayValue
	case commontest.ObjectValue:
		return ObjectValue
	}
	return NullValue
}
