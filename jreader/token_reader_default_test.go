//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// isEasyJSON is used in tests to e.g. expect different allocation behavior depending
// on which backend is in use.
const isEasyJSON = false

// These tests pin the default tokenReader's behavior on inputs that the broader permutation
// suites do not reach: malformed numbers and strings, wrong tokens where a structural
// delimiter is required, and the interaction of each entry point with a pushed-back token.
// Expected errors are asserted exactly (type, message, value, and offset), since all of
// those are part of the reader's observable behavior.

// parkToken parses the next token and pushes it back, leaving it as the tokenReader's
// unread token. Null does exactly that for any token that is not a null (including
// returning an error for a lone punctuation token, which still leaves the token parked).
func parkToken(tr *tokenReader) {
	_, _ = tr.Null()
}

func TestTokenReaderNumberEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    float64
		wantErr error
	}{
		{name: "uppercase exponent marker", input: "2E3", want: 2000},
		{name: "negative zero", input: "-0", want: 0},
		{name: "leading zero accepted", input: "012", want: 12},
		{name: "negative leading zero accepted", input: "-01", want: -1},
		{name: "dot with no fractional digits accepted", input: "1.", want: 1},
		{name: "no integer part accepted", input: "-.123", want: -0.123},
		{name: "dot with no fractional digits before exponent accepted", input: "2.e3", want: 2000},
		{name: "integer beyond int64 wraps", input: "9223372036854775808",
			want: -9223372036854775808},
		{name: "exponent overflowing float64 range rejected", input: "1e400",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "exponent marker at end of input", input: "1e",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "exponent marker followed by non-digit", input: "1ex",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "exponent sign at end of input", input: "1e+",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "exponent sign followed by non-digit", input: "1e-x",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "minus sign alone", input: "-",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "error offset counts leading whitespace", input: " 1e",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			n, err := tr.Number()
			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, n)
		})
	}
}

func TestTokenReaderStringDecodingEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "control-character escapes", input: `"\b\f\n\r"`, want: "\b\f\n\r"},
		{name: "uppercase hex digits in unicode escape", input: `"\u00AF"`, want: "\u00af"},
		{name: "multi-byte character after an escape", input: "\"\\t\u00e9\"", want: "\t\u00e9"},
		{name: "invalid UTF-8 byte after an escape becomes replacement character",
			input: "\"\\t\xffx\"", want: "\t\ufffdx"},
		{name: "raw NUL character passes through", input: "\"a\x00a\"", want: "a\x00a"},
		{name: "raw tab passes through", input: "\"\t\"", want: "\t"},
		{name: "raw newline passes through", input: "\"new\nline\"", want: "new\nline"},
		{name: "unterminated with no escape", input: `"abc`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		{name: "unterminated after an escape", input: `"a\tb`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		{name: "backslash at end of input", input: `"a\`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		{name: "invalid escape character", input: `"\q"`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		{name: "non-hex digit in unicode escape", input: `"\u12G4"`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		{name: "unicode escape truncated by end of input", input: `"\u12`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			s, err := tr.String()
			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, s)
		})
	}
}

func TestTokenReaderPropertyNameEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "name followed by colon", input: `"a":`, want: "a"},
		{name: "input ends after name", input: `"a"`, wantErr: io.EOF},
		{name: "value instead of colon", input: `"a" 1`,
			wantErr: SyntaxError{Message: errMsgExpectedColon, Value: "number", Offset: 4}},
		{name: "malformed token instead of colon", input: `"a" tru`,
			wantErr: SyntaxError{Message: errMsgUnexpectedSymbol, Value: "tru", Offset: 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			name, err := tr.PropertyName()
			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, string(name))
		})
	}
}

func TestTokenReaderEndDelimiterOrCommaEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     string
		park      bool // parse and push back the first token before the call under test
		delimiter byte
		wantEnd   bool
		wantErr   error
	}{
		{name: "pushed-back end delimiter matches", input: "]", park: true, delimiter: ']',
			wantEnd: true},
		{name: "pushed-back comma continues", input: ",", park: true, delimiter: ']'},
		{name: "pushed-back value token in array", input: "5", park: true, delimiter: ']',
			wantErr: SyntaxError{Message: errMsgBadArrayItem, Value: "number", Offset: 0}},
		{name: "pushed-back value token in object", input: "5", park: true, delimiter: '}',
			wantErr: SyntaxError{Message: errMsgBadObjectItem, Value: "number", Offset: 0}},
		{name: "pushed-back wrong punctuation", input: ":", park: true, delimiter: ']',
			wantErr: SyntaxError{Message: errMsgBadArrayItem, Value: "':'", Offset: 0}},
		{name: "end of input", input: "", delimiter: ']', wantErr: io.EOF},
		{name: "value token", input: "5", delimiter: ']',
			wantErr: SyntaxError{Message: errMsgBadArrayItem, Value: "number", Offset: 0}},
		{name: "array start where an array should end", input: "[", delimiter: ']',
			wantErr: SyntaxError{Message: errMsgBadArrayItem, Value: "array", Offset: 0}},
		{name: "malformed token", input: "tru", delimiter: ']',
			wantErr: SyntaxError{Message: errMsgUnexpectedSymbol, Value: "tru", Offset: 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			if tc.park {
				parkToken(&tr)
			}
			isEnd, err := tr.EndDelimiterOrComma(tc.delimiter)
			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantEnd, isEnd)
		})
	}
}

func TestTokenReaderAnyEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		park    bool
		want    AnyValue
		wantErr error
	}{
		{name: "pushed-back array start", input: "[", park: true,
			want: AnyValue{Kind: ArrayValue}},
		{name: "pushed-back object start", input: "{", park: true,
			want: AnyValue{Kind: ObjectValue}},
		{name: "pushed-back number", input: "5", park: true,
			want: AnyValue{Kind: NumberValue, Number: 5}},
		{name: "pushed-back punctuation", input: ":", park: true,
			wantErr: SyntaxError{Message: errMsgUnexpectedChar, Value: ":", Offset: 0}},
		{name: "malformed number", input: "1e",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "unterminated string", input: `"abc`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			if tc.park {
				parkToken(&tr)
			}
			got, err := tr.Any()
			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestTokenReaderNextTokenErrors(t *testing.T) {
	// Null is the generic next()-driven entry point: any malformed token surfaces through it.
	for _, tc := range []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "malformed number", input: "1e",
			wantErr: SyntaxError{Message: errMsgInvalidNumber, Offset: 0}},
		{name: "unterminated string", input: `"abc`,
			wantErr: SyntaxError{Message: errMsgInvalidString, Offset: 0}},
		// The offending byte is reported as a code point (string(byte) converts as a rune),
		// so 0xEF renders as U+00EF.
		{name: "UTF-8 byte order mark", input: "\xef\xbb\xbf{}",
			wantErr: SyntaxError{Message: errMsgUnexpectedChar, Value: "ï", Offset: 0}},
		{name: "UTF-16 byte order mark", input: "\xff\xfe{\x00}\x00",
			wantErr: SyntaxError{Message: errMsgUnexpectedChar, Value: "ÿ", Offset: 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte(tc.input))
			_, err := tr.Null()
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestTokenReaderNonASCIIWhitespaceBytes(t *testing.T) {
	// The whitespace table classifies each byte the way unicode.IsSpace classifies its
	// Latin-1 code point, so these single bytes are skipped between tokens along with the
	// standard space, tab, CR, and LF.
	for _, tc := range []struct {
		name string
		ws   byte
	}{
		{name: "vertical tab", ws: 0x0B},
		{name: "form feed", ws: 0x0C},
		{name: "next line", ws: 0x85},
		{name: "no-break space", ws: 0xA0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTokenReader([]byte{tc.ws, '5', tc.ws})
			n, err := tr.Number()
			require.NoError(t, err)
			require.Equal(t, float64(5), n)
			require.Equal(t, 1, tr.LastPos())
			require.True(t, tr.EOF())
		})
	}
}

func TestTokenReaderDeepNestingNeedsNoPerDepthState(t *testing.T) {
	// The tokenizer tracks no nesting state (matching brackets is the caller's job), so
	// arbitrarily deep structures scan in constant space with no recursion.
	const depth = 100_000
	data := make([]byte, 0, 2*depth)
	for i := 0; i < depth; i++ {
		data = append(data, '[')
	}
	for i := 0; i < depth; i++ {
		data = append(data, ']')
	}
	tr := newTokenReader(data)
	for i := 0; i < depth; i++ {
		found, err := tr.Delimiter('[')
		require.NoError(t, err)
		require.True(t, found)
	}
	for i := 0; i < depth; i++ {
		isEnd, err := tr.EndDelimiterOrComma(']')
		require.NoError(t, err)
		require.True(t, isEnd)
	}
	require.True(t, tr.EOF())
}

func TestTokenReaderEOFWithPushedBackToken(t *testing.T) {
	tr := newTokenReader([]byte("false"))
	parkToken(&tr)
	require.False(t, tr.EOF(), "a pushed-back token means the input is not exhausted")
	b, err := tr.Bool()
	require.NoError(t, err)
	require.False(t, b)
	require.True(t, tr.EOF())
}

func TestTokenReaderDelimiterAtEndOfInput(t *testing.T) {
	tr := newTokenReader([]byte("  "))
	found, err := tr.Delimiter('[')
	require.NoError(t, err)
	require.False(t, found)
}

// The remaining tests cover defensive branches of internal helpers directly, because no
// input can reach them through the entry points: getPos is only called after a putBack,
// readNumber always passes parseIntFromBytes at least one byte, and valueKind filters out
// the delimiter kinds before calling valueKindFromTokenKind.

func TestTokenReaderGetPosWithoutPushedBackToken(t *testing.T) {
	tr := newTokenReader([]byte(" 5 "))
	_, err := tr.Number()
	require.NoError(t, err)
	require.Equal(t, 2, tr.getPos())
}

func TestParseIntFromBytesRejectsEmptyInput(t *testing.T) {
	_, ok := parseIntFromBytes(nil)
	require.False(t, ok)
}

func TestValueKindFromTokenKindHasNoDelimiterMapping(t *testing.T) {
	require.Equal(t, ValueKind(-1), valueKindFromTokenKind(delimiterToken))
}
