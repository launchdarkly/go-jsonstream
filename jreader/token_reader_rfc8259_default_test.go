package jreader

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise RFC 8259 compliance of the tokenizer.
//
// The authority is encoding/json: json.Valid for grammar (accept/reject) and json.Unmarshal
// for decoded values. One deliberate exception: numeric literals whose magnitude exceeds
// float64's range (e.g. 1e400) are valid per the JSON grammar but cannot be represented, so --
// like json.Unmarshal into a float64 or interface{} -- the reader rejects them.

// rejectedByEncodingJSON is a sanity check so the "must reject" table cannot drift into
// asserting that valid JSON is rejected.
func rejectedByEncodingJSON(t *testing.T, input string) {
	t.Helper()
	require.False(t, json.Valid([]byte(input)),
		"test bug: %q is actually valid JSON, so the reader should not reject it", input)
}

func parseWholeValue(input string) error {
	r := NewReader([]byte(input))
	r.Any()
	if err := r.Error(); err != nil {
		return err
	}
	return r.RequireEOF()
}

func TestReaderRejectsMalformedWhitespace(t *testing.T) {
	// JSON permits only space, tab, LF, and CR as whitespace between tokens.
	for _, input := range []string{
		"\x0b1",     // vertical tab
		"\x0c1",     // form feed
		"1\x0b",     // vertical tab as trailing content
		"[\x0c1]",   // form feed inside a structure
		"\xc2\x851", // NEL (U+0085), encoded as UTF-8
		"\xc2\xa01", // NBSP (U+00A0), encoded as UTF-8
	} {
		t.Run(input, func(t *testing.T) {
			rejectedByEncodingJSON(t, input)
			require.Error(t, parseWholeValue(input))
		})
	}
}

func TestReaderAcceptsAllJSONWhitespace(t *testing.T) {
	// Carriage return in particular was not covered by the shared suite.
	for _, input := range []string{
		" 1 ",
		"\t1\t",
		"\n1\n",
		"\r1\r",
		"\r\n 1 \r\n",
	} {
		t.Run(input, func(t *testing.T) {
			require.True(t, json.Valid([]byte(input)))
			require.NoError(t, parseWholeValue(input))
		})
	}
}

func TestReaderRejectsMalformedNumbers(t *testing.T) {
	for _, input := range []string{
		"01",
		"-01",
		"00",
		"007",
		"1.",
		"-1.",
		"1.e3",
		"-",
		"1e",
		"1e+",
		"1E-",
		"1.2.3",
	} {
		t.Run(input, func(t *testing.T) {
			rejectedByEncodingJSON(t, input)
			require.Error(t, parseWholeValue(input))
		})
	}
}

func TestReaderAcceptsValidNumbers(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  float64
	}{
		{"0", 0},
		{"-0", 0},
		{"3", 3},
		{"-3", 3 * -1},
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"0.5", 0.5},
		{"1e3", 1000},
		{"1E3", 1000},
		{"1e+3", 1000},
		{"1e-3", 0.001},
		{"1.5e2", 150},
		{"10", 10},
		{"100", 100},
		{"0e1", 0},
		{"0.0e0", 0},
		{"1e10", 1e10},
		{"1E+2", 100},
	} {
		t.Run(tc.input, func(t *testing.T) {
			require.True(t, json.Valid([]byte(tc.input)))
			r := NewReader([]byte(tc.input))
			got := r.Float64()
			require.NoError(t, r.Error())
			require.NoError(t, r.RequireEOF())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReaderRejectsUnescapedControlCharsInString(t *testing.T) {
	for _, input := range []string{
		"\"a\tb\"", // literal tab
		"\"a\nb\"", // literal newline
		"\"a\rb\"", // literal carriage return
		"\"\x00\"", // NUL
		"\"\x01\"", // SOH
		"\"\x1f\"", // unit separator (highest control char)
	} {
		t.Run(input, func(t *testing.T) {
			rejectedByEncodingJSON(t, input)
			require.Error(t, parseWholeValue(input))
		})
	}
}

func TestReaderCombinesSurrogatePairs(t *testing.T) {
	// A \u-escaped UTF-16 surrogate pair must decode to a single code point rather than two
	// replacement characters. These use raw string literals (backticks) so the literal
	// backslash-u escape bytes reach the reader -- and readUnicodeEscape's combining branch --
	// rather than being resolved to decoded text by the Go compiler.
	for _, tc := range []struct {
		input string
		want  string
	}{
		{`"\uD834\uDD1E"`, "\U0001D11E"},     // musical G clef
		{`"a\uD834\uDD1Eb"`, "a\U0001D11Eb"}, // with surrounding characters
		{`"\uD83D\uDE02"`, "\U0001F602"},     // face with tears of joy
		{`"\uDBFF\uDFFF"`, "\U0010FFFF"},     // maximum code point
	} {
		t.Run(tc.input, func(t *testing.T) {
			var stdlib string
			require.NoError(t, json.Unmarshal([]byte(tc.input), &stdlib))
			assert.Equal(t, tc.want, stdlib, "test bug: expectation disagrees with encoding/json")

			r := NewReader([]byte(tc.input))
			got := r.String()
			require.NoError(t, r.Error())
			require.NoError(t, r.RequireEOF())
			assert.Equal(t, tc.want, got)
		})
	}

	// The same code point encoded as literal UTF-8 bytes must also decode intact -- this goes
	// through the plain ReadRune path rather than the escape path.
	r := NewReader([]byte(`"a𝄞b"`))
	got := r.String()
	require.NoError(t, r.Error())
	require.NoError(t, r.RequireEOF())
	assert.Equal(t, "a\U0001D11Eb", got)
}

func TestReaderReplacesInvalidSurrogates(t *testing.T) {
	// Lone or malformed surrogates decode to the replacement character, matching
	// encoding/json (which does not treat them as an error).
	for _, tc := range []struct {
		input string
		want  string
	}{
		{`"\uD834"`, "�"},        // lone high surrogate
		{`"\uDD1E"`, "�"},        // lone low surrogate
		{`"\uD834x"`, "�x"},      // high surrogate followed by a normal char
		{`"\uD834\uD834"`, "��"}, // high surrogate followed by another high surrogate
	} {
		t.Run(tc.input, func(t *testing.T) {
			var stdlib string
			require.NoError(t, json.Unmarshal([]byte(tc.input), &stdlib))
			assert.Equal(t, tc.want, stdlib, "test bug: expectation disagrees with encoding/json")

			r := NewReader([]byte(tc.input))
			got := r.String()
			require.NoError(t, r.Error())
			require.NoError(t, r.RequireEOF())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReaderRejectsControlCharInPropertyName(t *testing.T) {
	// The control-char rule applies to property names, not just string values.
	input := "{\"a\tb\":1}"
	require.False(t, json.Valid([]byte(input)))

	r := NewReader([]byte(input))
	obj := r.Object()
	obj.Next()
	require.Error(t, r.Error())
}

func TestReaderSubstitutesInvalidUTF8(t *testing.T) {
	// Raw invalid UTF-8 bytes inside a string decode to the Unicode replacement character,
	// matching encoding/json, regardless of where an escape appears in the same string.
	for _, tc := range []struct {
		name  string
		input []byte
	}{
		{"lone invalid byte", []byte("\"\xff\"")},
		{"invalid byte then text", []byte("\"\xffab\"")},
		{"text then invalid byte", []byte("\"ab\xff\"")},
		{"invalid byte after escape", []byte("\"\\n\xff\"")},
		{"invalid byte before escape", []byte("\"\xff\\n\"")},
		{"CESU-8 encoded surrogate", []byte("\"\xed\xa0\x80\"")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, json.Valid(tc.input), "precondition: valid JSON grammar")
			var want string
			require.NoError(t, json.Unmarshal(tc.input, &want))

			r := NewReader(tc.input)
			got := r.String()
			require.NoError(t, r.Error())
			require.NoError(t, r.RequireEOF())
			assert.Equal(t, want, got)
		})
	}
}

func TestReaderLargeIntegersMatchEncodingJSON(t *testing.T) {
	// Integer literals that overflow int64 but are within float64 range must decode to the same
	// value encoding/json produces, not a wrapped-int64 garbage value.
	for _, input := range []string{
		"99999999999999999999",           // > 2^63
		"12345678901234567890",           // > 2^63
		"9223372036854775808",            // 2^63 exactly
		"18446744073709551616",           // 2^64
		"123456789012345678901234567890", // ~1.2e29, still within float64 range
	} {
		t.Run(input, func(t *testing.T) {
			var want float64
			require.NoError(t, json.Unmarshal([]byte(input), &want))

			r := NewReader([]byte(input))
			got := r.Float64()
			require.NoError(t, r.Error())
			require.NoError(t, r.RequireEOF())
			assert.Equal(t, want, got)
		})
	}
}

func TestReaderRejectsOutOfRangeNumbers(t *testing.T) {
	// Numbers whose magnitude exceeds float64's range are valid JSON grammar but cannot be
	// represented. encoding/json rejects them when decoding into a float64/interface{}, and so
	// does the reader (a deliberate exception to strict json.Valid parity).
	for _, input := range []string{
		"1e400",
		"9e999",
		"-1e400",
		"1E4000",                       // out-of-range float
		"1" + strings.Repeat("0", 400), // out-of-range integer literal
	} {
		t.Run(input, func(t *testing.T) {
			require.True(t, json.Valid([]byte(input)), "precondition: valid JSON grammar")
			var sink interface{}
			require.Error(t, json.Unmarshal([]byte(input), &sink),
				"precondition: encoding/json also rejects the value")

			require.Error(t, parseWholeValue(input))
		})
	}
}
