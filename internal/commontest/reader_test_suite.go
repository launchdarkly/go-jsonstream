package commontest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ReaderTestSuite runs a standard set of tests against some implementation of JSON reading.
// This allows us to test both jreader.Reader and the low-level tokenizer jreader.tokenReader
// with many permutations of valid and invalid input.
type ReaderTestSuite struct {
	// ContextFactory must be provided by the caller to create an implementation of TestContext for
	// running a parsing test on the specified JSON input. This should include whatever parser
	// object will be used by the Actions that the ValueTestFactory creates.
	ContextFactory func(input []byte) TestContext

	// ValueTestFactory must be provided by the caller to create implementations of Action for
	// various JSON value types.
	ValueTestFactory ValueTestFactory

	// ReadErrorTestFactory must be provided by the caller to define expectations about error
	// reporting for invalid input.
	ReadErrorTestFactory ReadErrorTestFactory
}

// Run runs the test suite.
func (s ReaderTestSuite) Run(t *testing.T) {
	tf := testFactory{
		valueTestFactory:     s.ValueTestFactory,
		readErrorTestFactory: s.ReadErrorTestFactory,
		encodingBehavior: encodingBehavior{
			forParsing: true,
		},
	}
	var testDefs testDefs
	testDefs = append(testDefs, tf.MakeAllValueTests()...)
	testDefs = append(testDefs, tf.MakeAllReadErrorTests()...)
	whitespaceOptions := MakeWhitespaceOptions()
	whitespaceOptions[""] = ""
	for _, td := range testDefs {
		for wsName, wsValue := range whitespaceOptions {
			testName := td.name
			if wsName != "" {
				testName += " [with whitespace: " + wsName + "]"
			}
			t.Run(testName, func(t *testing.T) {
				input := wsValue + strings.Join(td.encoding, wsValue) + wsValue
				t.Cleanup(func() {
					if t.Failed() {
						t.Logf("JSON input was: `%s`", input)
					}
				})
				c := s.ContextFactory([]byte(input))
				require.NoError(t, td.action(c))
			})
		}
	}
}
