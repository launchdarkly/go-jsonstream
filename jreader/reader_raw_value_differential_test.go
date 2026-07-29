//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests validate RawValue's accept/reject decisions and captured byte-span against
// encoding/json as an independent oracle, rather than against hardcoded expectations. The oracle
// for the span of the first JSON value is encoding/json's streaming decoder, which trims leading
// whitespace and stops at the end of the first value -- the same contract RawValue provides.

// oracleFirstValue returns the raw bytes of the first JSON value in input (leading whitespace
// trimmed, trailing content ignored), or ok == false if encoding/json rejects the input.
func oracleFirstValue(input []byte) (raw []byte, ok bool) {
	dec := json.NewDecoder(bytes.NewReader(input))
	var rm json.RawMessage
	if err := dec.Decode(&rm); err != nil {
		return nil, false
	}
	return []byte(rm), true
}

func TestRawValueBoundaryMatchesEncodingJSON(t *testing.T) {
	inputs := []string{
		// scalars
		`0`, `123`, `-456`, `3.25`, `-3.25e2`, `1e5`, `1E+5`, `0.0`, `-0`,
		`true`, `false`, `null`,
		`""`, `"simple"`, `"with \"escaped\" quotes"`, `"\\"`, `"\""`, `"A"`, `"multi é byte"`,
		// scalar followed by trailing content -- the span must stop at the value boundary
		`123 `, `123,`, `123}`, `123]`, `123   trailing`, `"x"extra`, `true false`,
		// containers
		`{}`, `[]`, `{"a":1}`, `[1,2,3]`,
		`{"a":{"b":[{"c":null}]},"d":"e"}`,
		`["nested",["arrays",{"and":"objects"}]]`,
		`{"tricky":"a \"}\" and ] inside","escaped\"name":1}`,
		// containers with trailing junk
		`{}extra`, `[1,2]  trailing`, `[1,2,3]}`, `{"a":1},`,
		// strings containing bracket-like bytes that must not confuse the scan
		`["]"]`, `["}"]`, `{"k]":"v"}`, `{"k[":"v"}`, `["\\"]`, `["\""]`, `["a\\","b"]`,
		// adversarial escapes: escaped backslash/quote boundaries and an escaped ']' that must
		// stay inside the string
		`["\\\\"]`, `["a\\"]`, `{"\\":"\\"}`, `["\u005D"]`, `["\\}"]`, `{"a":"\\","b":["\""]}`,
		// invalid inputs -- both the oracle and RawValue must reject these
		`["\}"]`, `["\]"]`, `[1 2]`, `{"a":}`, `[1,2,,]`, `{"a":1,}`,
		// whitespace variations
		`   {"a":1}   `, "\t\n [1] \n", ` "s" `,
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			want, ok := oracleFirstValue([]byte(in))
			r := NewReader([]byte(in))
			got := r.RawValue()
			if !ok {
				assert.Error(t, r.Error(), "encoding/json rejected %q but RawValue accepted %q", in, string(got))
				return
			}
			require.NoError(t, r.Error(), "encoding/json accepted %q as %q but RawValue errored", in, string(want))
			assert.Equal(t, string(want), string(got), "boundary mismatch for %q", in)
		})
	}
}

// A scalar whose valid prefix is itself a complete JSON value is captured up to that boundary
// without error; the trailing garbage is surfaced later via RequireEOF, not by RawValue.
func TestRawValueReturnsValidScalarPrefix(t *testing.T) {
	for _, c := range []struct{ input, want string }{
		{`1.2.3`, `1.2`},
		{`0x1`, `0`},
		{`123abc`, `123`},
		{`1e5e5`, `1e5`},
	} {
		t.Run(c.input, func(t *testing.T) {
			r := NewReader([]byte(c.input))
			raw := r.RawValue()
			require.NoError(t, r.Error(), "prefix %q wrongly rejected", c.input)
			assert.Equal(t, c.want, string(raw))
			assert.True(t, json.Valid(raw), "returned prefix must be valid JSON")
			assert.Error(t, r.RequireEOF(), "trailing garbage should surface via RequireEOF")
		})
	}
}

// genValue builds a random JSON-encodable value up to the given nesting depth, favoring content
// (brackets, quotes, escapes, multi-byte runes) that stresses the string-aware boundary scan.
func genValue(rng *rand.Rand, depth int) interface{} {
	if depth <= 0 || rng.Intn(3) == 0 {
		switch rng.Intn(6) {
		case 0:
			return nil
		case 1:
			return rng.Intn(2) == 0
		case 2:
			return rng.NormFloat64() * 1000
		case 3:
			return rng.Intn(1000000) - 500000
		default:
			pool := []string{"", "a", `q"q`, `back\slash`, "]}[{", "unicode é", "tab\tnl\n", `\/`}
			return pool[rng.Intn(len(pool))]
		}
	}
	if rng.Intn(2) == 0 {
		arr := make([]interface{}, rng.Intn(4))
		for i := range arr {
			arr[i] = genValue(rng, depth-1)
		}
		return arr
	}
	m := map[string]interface{}{}
	keys := []string{"a", "b", `k"x`, "]", "}", "nested", "é"}
	for i := 0; i < rng.Intn(4); i++ {
		m[keys[rng.Intn(len(keys))]] = genValue(rng, depth-1)
	}
	return m
}

// The captured span must match encoding/json's first-value span for thousands of random values,
// including random trailing bytes after the value, and must always be a cap==len alias.
func TestRawValueGenerativeDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	trailers := []string{"", " ", ",", "]", "}", "  junk", `"next"`, "42", "\t\n"}
	for iter := 0; iter < 20000; iter++ {
		canonical, err := json.Marshal(genValue(rng, 4))
		if err != nil {
			continue
		}
		input := append(append([]byte(nil), canonical...), trailers[rng.Intn(len(trailers))]...)

		want, ok := oracleFirstValue(input)
		if !ok {
			continue // ambiguous trailing (e.g. a number run-on); skip
		}
		r := NewReader(append([]byte(nil), input...))
		got := r.RawValue()
		require.NoError(t, r.Error(), "iter %d: RawValue errored on %q (oracle=%q)", iter, string(input), string(want))
		require.Equal(t, string(want), string(got), "iter %d: boundary mismatch input=%q", iter, string(input))
		require.Equal(t, len(got), cap(got), "iter %d: cap!=len for %q", iter, string(input))
	}
}

// Fuzz the number grammar with random number-ish bytes and assert three properties that together
// rule out both invalid captures and false rejections of valid scalars.
func TestRawValueScalarGrammarFuzz(t *testing.T) {
	rng := rand.New(rand.NewSource(20260728))
	alphabet := []byte("0123456789+-.eE ")
	for i := 0; i < 50000; i++ {
		b := make([]byte, 1+rng.Intn(8))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		in := string(b)

		r := NewReader([]byte(in))
		raw := r.RawValue()
		if r.Error() != nil {
			continue
		}
		// Property A: anything returned must itself be valid JSON.
		require.True(t, json.Valid(raw), "RawValue returned non-valid-JSON bytes %q for input %q", raw, in)
		// Property B: the returned bytes must be a boundary-correct prefix of the trimmed input.
		trimmed := strings.TrimLeft(in, " ")
		require.True(t, strings.HasPrefix(trimmed, string(raw)),
			"RawValue %q is not a prefix of trimmed input %q", raw, trimmed)
		// Property C (no false rejection): if the whole trimmed input is a single valid scalar,
		// RawValue must have returned exactly it.
		if tr := strings.TrimSpace(in); tr != "" && json.Valid([]byte(tr)) && !strings.ContainsAny(tr, "{}[]") {
			require.Equal(t, tr, string(raw), "input %q is a single valid scalar but RawValue returned %q", tr, raw)
		}
	}
}
