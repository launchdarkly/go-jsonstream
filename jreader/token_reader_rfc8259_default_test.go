//go:build !launchdarkly_easyjson

package jreader

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise RFC 8259 compliance of the default tokenizer. They are default-only
// because the easyjson backend does not make all of the same accept/reject decisions. The
// authority for whether an input is well-formed JSON is encoding/json's json.Valid.

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
	// A valid UTF-16 surrogate pair must decode to a single code point rather than two
	// replacement characters. U+1D11E (musical G clef) is 𝄞.
	r := NewReader([]byte(`"𝄞"`))
	got := r.String()
	require.NoError(t, r.Error())
	require.NoError(t, r.RequireEOF())
	assert.Equal(t, "\U0001D11E", got)

	// Surrounding characters are preserved.
	r = NewReader([]byte(`"a𝄞b"`))
	got = r.String()
	require.NoError(t, r.Error())
	assert.Equal(t, "a\U0001D11Eb", got)
}

func TestReaderReplacesInvalidSurrogates(t *testing.T) {
	// Lone or malformed surrogates decode to the replacement character, matching
	// encoding/json (which does not treat them as an error).
	for _, tc := range []struct {
		input string
		want  string
	}{
		{`"\uD834"`, "�"},              // lone high surrogate
		{`"\uDD1E"`, "�"},              // lone low surrogate
		{`"\uD834x"`, "�x"},            // high surrogate followed by a normal char
		{`"\uD834\uD834"`, "��"},  // high surrogate followed by another high surrogate
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
