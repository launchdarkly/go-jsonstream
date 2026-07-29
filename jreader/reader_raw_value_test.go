//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RawValue is provided only by the default build; the launchdarkly_easyjson build has no such
// method (easyjson support is planned for removal, so no new functionality is provided for it).
// All of its tests therefore carry the !launchdarkly_easyjson tag.

func TestRawValueScalars(t *testing.T) {
	for _, input := range []string{
		`null`,
		`true`,
		`false`,
		`0`,
		`123`,
		`-456`,
		`3.25`,
		`-3.25e2`,
		`""`,
		`"simple"`,
		`"with \"escaped\" quotes"`,
		`"unicode é escape"`,
		`"multi-byte é directly"`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			raw := r.RawValue()
			require.NoError(t, r.Error())
			assert.Equal(t, input, string(raw))
			assert.NoError(t, r.RequireEOF())
		})
	}
}

func TestRawValueContainers(t *testing.T) {
	for _, input := range []string{
		`{}`,
		`[]`,
		`{"a":1}`,
		`[1,2,3]`,
		`{"a":{"b":[{"c":null}]},"d":"e"}`,
		`["nested",["arrays",{"and":"objects"}]]`,
		`{"tricky":"a \"}\" and ] inside a string","escaped\"name":1}`,
		`{ "spaced" : [ 1 , 2 ] }`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			raw := r.RawValue()
			require.NoError(t, r.Error())
			assert.Equal(t, input, string(raw))
			assert.NoError(t, r.RequireEOF())
		})
	}
}

func TestRawValueSkipsLeadingWhitespaceAndLeavesTrailingInput(t *testing.T) {
	r := NewReader([]byte(`   {"a":1}   `))
	raw := r.RawValue()
	require.NoError(t, r.Error())
	assert.Equal(t, `{"a":1}`, string(raw))
	assert.NoError(t, r.RequireEOF())
}

func TestRawValueWithinObjectIteration(t *testing.T) {
	input := `{"first":true,"raw":{"x":[1,2],"y":"z"},"last":3}`
	r := NewReader([]byte(input))
	var raw []byte
	var first bool
	var last float64
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case "first":
			first = r.Bool()
		case "raw":
			raw = r.RawValue()
		case "last":
			last = r.Float64()
		}
	}
	require.NoError(t, r.Error())
	assert.True(t, first)
	assert.Equal(t, `{"x":[1,2],"y":"z"}`, string(raw))
	assert.Equal(t, float64(3), last)
	assert.NoError(t, r.RequireEOF())
}

func TestRawValueWithinArrayIteration(t *testing.T) {
	input := `[10,{"a":"b"},[true],20]`
	r := NewReader([]byte(input))
	var values []string
	for arr := r.Array(); arr.Next(); {
		values = append(values, string(r.RawValue()))
	}
	require.NoError(t, r.Error())
	assert.Equal(t, []string{`10`, `{"a":"b"}`, `[true]`, `20`}, values)
	assert.NoError(t, r.RequireEOF())
}

func TestRawValueEveryValueTypeWithinObjectIteration(t *testing.T) {
	input := `{"null":null,"bool":true,"num":-1.5,"str":"s","arr":[],"obj":{}}`
	expected := map[string]string{
		"null": `null`, "bool": `true`, "num": `-1.5`, "str": `"s"`, "arr": `[]`, "obj": `{}`,
	}
	r := NewReader([]byte(input))
	got := map[string]string{}
	for obj := r.Object(); obj.Next(); {
		got[string(obj.Name())] = string(r.RawValue())
	}
	require.NoError(t, r.Error())
	assert.Equal(t, expected, got)
}

func TestRawValueRoundTripsThroughReader(t *testing.T) {
	input := `{"a":[1,{"b":"c \" d"}],"e":null}`
	r := NewReader([]byte(input))
	raw := r.RawValue()
	require.NoError(t, r.Error())

	// The captured bytes must themselves be a complete, parseable JSON value.
	r2 := NewReader(raw)
	require.NoError(t, r2.SkipValue())
	require.NoError(t, r2.Error())
	assert.NoError(t, r2.RequireEOF())
}

func TestRawValueOnEmptyInputReturnsError(t *testing.T) {
	r := NewReader(nil)
	raw := r.RawValue()
	assert.Nil(t, raw)
	assert.Error(t, r.Error())
}

func TestRawValueOnTruncatedContainerReturnsError(t *testing.T) {
	for _, input := range []string{
		`{"a":`,
		`{"a":1`,
		`[1,2`,
		`{"a":"unterminated`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			_ = r.RawValue()
			assert.Error(t, r.Error())
		})
	}
}

func TestRawValueAfterPreviousErrorReturnsNil(t *testing.T) {
	r := NewReader([]byte(`{"a":1}`))
	r.AddError(SyntaxError{Message: "sorry"})
	assert.Nil(t, r.RawValue())
	assert.Error(t, r.Error())
}

// Containers with mismatched bracket kinds must be rejected rather than captured as truncated or
// structurally broken bytes.
func TestRawValueRejectsMalformedContainer(t *testing.T) {
	for _, input := range []string{
		`{"a":1]`,
		`[1,2}`,
		`{"a":[1}]`,
		`[}]`,
		`{"a":]}`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			_ = r.RawValue()
			assert.Error(t, r.Error())
		})
	}
}

// The captured bytes must be a zero-copy view into the input, not a copy. Downstream relaying (the
// FDv2 use case) depends on this, and the returned slice must have cap == len so a caller that
// appends to or reslices it cannot reach adjacent bytes of the input buffer.
func TestRawValueAliasesInputWithoutExtraCapacity(t *testing.T) {
	for _, input := range []string{`{"a":1}`, `[1,2,3]`, `"scalar"`, `12345`} {
		t.Run(input, func(t *testing.T) {
			buf := []byte(input)
			r := NewReader(buf)
			raw := r.RawValue()
			require.NoError(t, r.Error())
			require.NotEmpty(t, raw)

			assert.Equal(t, len(raw), cap(raw), "cap must equal len")

			// Mutating the input must be visible through raw, proving they share a backing array.
			buf[0] = 'Z'
			assert.Equal(t, byte('Z'), raw[0], "raw must alias the input buffer")
		})
	}
}

// Nesting deeper than the scanner's inline capacity must round-trip exactly.
func TestRawValueDeepNesting(t *testing.T) {
	const depth = 40
	input := strings.Repeat("[", depth) + strings.Repeat("]", depth)
	r := NewReader([]byte(input))
	raw := r.RawValue()
	require.NoError(t, r.Error())
	assert.Equal(t, input, string(raw))
	assert.NoError(t, r.RequireEOF())
}

// A captured array or object is fully validated with encoding/json, so structurally bracketed
// but otherwise malformed content is rejected.
func TestRawValueRejectsInvalidContainerContent(t *testing.T) {
	for _, input := range []string{
		`[1 2]`,
		`{"a":}`,
		`[1,2,,,]`,
		`{"a" "b"}`,
		`{"a":1,}`,
		`["\x"]`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			_ = r.RawValue()
			assert.Error(t, r.Error())
		})
	}
}

func TestRawValueRejectsNonValueDelimiter(t *testing.T) {
	for _, input := range []string{
		`}`,
		`]`,
		`,`,
		`:`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			_ = r.RawValue()
			assert.Error(t, r.Error())
		})
	}
}

func TestRawValueRejectsMalformedScalar(t *testing.T) {
	for _, input := range []string{
		`tru`,
		`nul`,
		`-`,
		`1e`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			_ = r.RawValue()
			assert.Error(t, r.Error())
		})
	}
}

// Numbers that are not valid standalone JSON (leading zeros, a trailing decimal point, an empty
// exponent) must be rejected. The RFC 8259 compliant tokenizer rejects these itself; this test
// pins the RawValue-level guarantee regardless of which layer enforces it.
func TestRawValueRejectsScalarInvalidAsJSON(t *testing.T) {
	for _, input := range []string{
		`01`,
		`00`,
		`007`,
		`-0123`,
		`1.`,
		`1.e5`,
		`-.5`,
	} {
		t.Run(input, func(t *testing.T) {
			r := NewReader([]byte(input))
			raw := r.RawValue()
			assert.Nil(t, raw)
			assert.Error(t, r.Error())
		})
	}
}

// An invalid scalar can never enter the unread-token state, because the RFC 8259 compliant
// tokenizer rejects it at the initial probe; the rejection happens no later than the first
// attempt to read the token.
func TestRawValueRejectsScalarInvalidAsJSONAfterUnread(t *testing.T) {
	for _, input := range []string{
		`01`,
		`1.`,
		`1.e5`,
	} {
		t.Run(input, func(t *testing.T) {
			tr := newTokenReader([]byte(input))
			_, err := tr.Null()
			if err == nil {
				_, err = tr.RawValue()
			}
			assert.Error(t, err)
		})
	}
}

func TestTokenReaderRawValueAfterUnreadToken(t *testing.T) {
	// Probing the tokenizer with Null() parses the next token and pushes it back when it is not
	// a null, so RawValue must handle the unread-token state for every value type.
	for _, input := range []string{
		`true`,
		`123`,
		`"str"`,
		`{"a":[1,2]}`,
		`[{"b":"c"}]`,
	} {
		t.Run(input, func(t *testing.T) {
			tr := newTokenReader([]byte(input))
			isNull, err := tr.Null()
			require.NoError(t, err)
			require.False(t, isNull)
			raw, err := tr.RawValue()
			require.NoError(t, err)
			assert.Equal(t, input, string(raw))
			assert.True(t, tr.EOF())
		})
	}
}

func TestTokenReaderRawValueAfterUnreadNonValueDelimiterReturnsError(t *testing.T) {
	// A failed Delimiter probe leaves the mismatched delimiter token unread.
	tr := newTokenReader([]byte(`}`))
	gotDelim, err := tr.Delimiter(']')
	require.NoError(t, err)
	require.False(t, gotDelim)
	_, err = tr.RawValue()
	assert.Error(t, err)
}
