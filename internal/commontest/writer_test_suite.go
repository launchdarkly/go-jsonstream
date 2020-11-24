package commontest

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// WriterTestSuite runs a standard set of tests against some implementation of JSON writing.
// This allows us to test both jwriter.Writer and the low-level JSON formatter jwriter.tokenWriter
// with many permutations of output data.
type WriterTestSuite struct {
	// ContextFactory must be provided by the caller to create an implementation of TestContext for
	// running a writing test on some set of JSON data. This should include whatever writer object
	// will be used by the Actions that the ValueTestFactory creates.
	ContextFactory func() TestContext

	// ValueTestFactory must be provided by the caller to create implementations of Action for
	// various JSON value types.
	ValueTestFactory ValueTestFactory

	// EncodeAsHex must be provided by the caller to define expectations about whether this writer
	// will use a \uNNNN escape sequence for the specified Unicode character. There is no single
	// correct answer for all implementations, since JSON allows characters to be escaped in
	// several ways and also allows unescaped multi-byte characters.
	EncodeAsHex func(rune) bool
}

// Run runs the test suite.
func (s WriterTestSuite) Run(t *testing.T) {
	tf := testFactory{
		valueTestFactory: s.ValueTestFactory,
		encodingBehavior: encodingBehavior{
			encodeAsHex: s.EncodeAsHex,
		},
	}
	tds := tf.MakeAllValueTests()
	for _, td := range tds {
		t.Run(td.name, func(t *testing.T) {
			c := s.ContextFactory()
			t.Cleanup(func() {
				if t.Failed() {
					t.Logf("JSON output: `%s`", string(c.JSONData()))
				}
			})
			td.action(c)
			output := string(c.JSONData())
			require.Regexp(t, makeOutputRegex(td.encoding), output)
		})
	}
}

// Make a regex that will allow any amount of whitespace between the matched substrings.
func makeOutputRegex(outputParts []string) *regexp.Regexp {
	regex := "\\w*"
	for _, outputPart := range outputParts {
		regex += regexp.QuoteMeta(outputPart)
		regex += "\\w*"
	}
	return regexp.MustCompile(regex)
}
