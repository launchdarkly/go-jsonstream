package commontest

// TestContext is an abstraction used by ReaderTestSuite and WriterTestSuite.
type TestContext interface {
	// JSONData returns either (for readers) the input data that was passed in when the TestContext
	// was created, or (for writers) all of the output that has been produced so far.
	JSONData() []byte
}

// Action is an arbitrary action that can be executed during a test. For readers, this normally
// consists of trying to read some type of value from the input, and asserting that no error occurred
// and that the expected value was found. For writers, it consists of trying to write something to
// the output.
//
// All test assertions should return early on any non-nil error.
type Action func(c TestContext) error

// PropertyAction is used in the context of a JSON object value, describing a property name and the
// Action for reading or writing the property value.
type PropertyAction struct {
	Name   string
	Action Action
}

// ValueVariant is an optional identifier that ValueTestFactory can use to make the tests produce
// multiple variations of value tests. See ValueTestFactory.Variants.
type ValueVariant string

const (
	// This variant means that the reader will try to consume a JSON value without regard to its type.
	ReadAsAnyTypeVariant ValueVariant = "any:"

	// This variant means that the reader will try to recursively skip past a JSON value of any type.
	SkipValueVariant ValueVariant = "skip:"
)

// ValueTestFactory is an interface for producing specific reader/writer test actions. To test any
// reader or writer with ReaderTestSuite or WriterTestSuite, provide an implementation of this
// interface that performs the specified actions.
type ValueTestFactory interface {
	EOF() Action
	Value(value AnyValue, variant ValueVariant) Action
	Variants(value AnyValue) []ValueVariant
}

type ReadErrorTestFactory interface {
	ExpectEOFError(err error) error
	ExpectWrongTypeError(err error, expectedType ValueKind, variant ValueVariant, gotType ValueKind) error
	ExpectSyntaxError(err error) error
}

type ValueKind int

const (
	NullValue   ValueKind = iota
	BoolValue   ValueKind = iota
	NumberValue ValueKind = iota
	StringValue ValueKind = iota
	ArrayValue  ValueKind = iota
	ObjectValue ValueKind = iota
)

type AnyValue struct {
	Kind   ValueKind
	Bool   bool
	Number float64
	String string
	Array  []Action
	Object []PropertyAction
}
