package commontest

import (
	"fmt"
	"strings"
)

// This file contains the logic for generating all the valid JSON permutations and error conditions
// that will be tested by ReaderTestSuite and WriterTestSuite.

type testFactory struct {
	valueTestFactory     ValueTestFactory
	readErrorTestFactory ReadErrorTestFactory
	encodingBehavior     encodingBehavior
}

// This struct associates a test Action with some JSON data.
type testDef struct {
	// A descriptive name, used in test output.
	name string

	// For readers, encoding is the JSON input data; for writers, the expected JSON output. It is
	// defined as a list of substrings, with the expectation that there may be any amount of
	// whitespace between each substring.
	encoding []string

	// The test action for either reading or writing this piece of JSON.
	action Action
}

type testDefs []testDef

func (td testDef) then(next testDef) testDef {
	return testDef{
		name:     td.name + ", " + next.name,
		encoding: append(td.encoding, next.encoding...),
		action: func(ctx TestContext) error {
			if err := td.action(ctx); err != nil {
				return err
			}
			return next.action(ctx)
		},
	}
}

func (tds testDefs) then(next testDef) testDefs {
	ret := make(testDefs, 0, len(tds))
	for _, td := range tds {
		ret = append(ret, td.then(next))
	}
	return ret
}

func (f testFactory) MakeAllValueTests() testDefs {
	ret := testDefs{}
	eofTest := testDef{name: "EOF", action: f.valueTestFactory.EOF()}
	ret = append(ret, f.makeScalarValueTests(true).then(eofTest)...)
	ret = append(ret, f.makeArrayTests().then(eofTest)...)
	ret = append(ret, f.makeObjectTests().then(eofTest)...)
	return ret
}

func (f testFactory) MakeAllReadErrorTests() testDefs {
	ret := testDefs{}
	addErrors := func(tds testDefs) {
		for _, td := range tds {
			for i, enc := range td.encoding {
				if enc == "" { // this means we want to force an unexpected EOF
					td.encoding = td.encoding[0:i]
					break
				}
			}
			ret = append(ret, td)
		}
	}
	addErrors(f.makeScalarValueReadErrorTests())
	return ret
}

func (f testFactory) makeScalarValueTests(allPermutations bool) testDefs {
	ret := testDefs{}
	values := f.makeScalarValues(allPermutations)
	oneVariant := []ValueVariant{""}
	for _, tv := range values {
		variants := f.valueTestFactory.Variants(tv.value)
		if variants == nil {
			variants = oneVariant
		}
		for _, variant := range variants {
			name := tv.name
			if variant != "" {
				name = string(variant) + " " + name
			}
			td := testDef{
				name:     name,
				encoding: []string{tv.encoding},
				action:   f.valueTestFactory.Value(tv.value, variant),
			}
			ret = append(ret, td)
		}
	}
	return ret
}

func (f testFactory) makeScalarValueReadErrorTests() testDefs {
	ret := testDefs{}
	values := f.makeScalarValues(false)
	oneVariant := []ValueVariant{""}
	for _, testValue := range values {
		tv := testValue
		variants := f.valueTestFactory.Variants(tv.value)
		if variants == nil {
			variants = oneVariant
		}
		for _, variant := range variants {
			v := variant
			name := tv.name
			if v != "" {
				name = string(v) + " " + name
			}
			testAction := f.valueTestFactory.Value(tv.value, variant)

			// error: want a value, got a value of some other type
			if v != ReadAsAnyTypeVariant && v != SkipValueVariant {
				for _, wrongValue := range f.makeScalarValues(false) {
					wv := wrongValue
					if wv.value.Kind == tv.value.Kind {
						continue
					}
					ret = append(ret, testDef{
						name:     fmt.Sprintf("%s (but got %s)", name, wv.name),
						encoding: []string{wv.encoding},
						action: func(c TestContext) error {
							return f.readErrorTestFactory.ExpectWrongTypeError(testAction(c), tv.value.Kind, v, wv.value.Kind)
						},
					})
				}
			}

			// error: want a value, got some invalid JSON
			for _, badThing := range []struct {
				name     string
				encoding string
			}{
				{"invalid identifier", "bad"},
				{"unknown delimiter", "+"},
				{"unexpected end array", "]"},
				{"unexpected object", "}"},
			} {
				ret = append(ret, testDef{
					name:     fmt.Sprintf("%s (but got %s)", name, badThing.name),
					encoding: []string{badThing.encoding},
					action: func(c TestContext) error {
						return f.readErrorTestFactory.ExpectSyntaxError(testAction(c))
					},
				})
			}
			ret = append(ret, testDef{
				name:     fmt.Sprintf("%s (but got unexpected EOF)", name),
				encoding: []string{""},
				action: func(c TestContext) error {
					return f.readErrorTestFactory.ExpectEOFError(testAction(c))
				},
			})
		}
	}
	return ret
}

func (f testFactory) makeScalarValues(allPermutations bool) []testValue {
	var values []testValue
	values = append(values, testValue{
		name:     "null",
		encoding: "null",
		value:    AnyValue{Kind: NullValue},
	})
	values = append(values, makeBoolTestValues()...)
	values = append(values, makeNumberTestValues(f.encodingBehavior)...)
	values = append(values, makeStringTestValues(f.encodingBehavior, allPermutations)...)
	return values
}

func (f testFactory) makeArrayTests() testDefs {
	ret := testDefs{}
	for elementCount := 0; elementCount <= 2; elementCount++ {
		for _, contents := range f.makeValueListsOfLength(elementCount) {
			var names []string
			var encoding = []string{"["}
			var actions []Action
			for i, td := range contents {
				names = append(names, td.name)
				if i > 0 {
					encoding = append(encoding, ",")
				}
				encoding = append(encoding, td.encoding...)
				actions = append(actions, td.action)
			}
			encoding = append(encoding, "]")
			value := AnyValue{Kind: ArrayValue, Array: actions}
			arrayTest := testDef{
				name:     "array(" + strings.Join(names, ", ") + ")",
				encoding: encoding,
				action:   f.valueTestFactory.Value(value, ""),
			}
			ret = append(ret, arrayTest)
		}
	}
	return ret
}

func (f testFactory) makeObjectTests() testDefs {
	ret := testDefs{}
	for propertyCount := 0; propertyCount <= 2; propertyCount++ {
		for _, contents := range f.makeValueListsOfLength(propertyCount) {
			var names []string
			var encoding = []string{"{"}
			var propActions []PropertyAction
			for i, td := range contents {
				propName := fmt.Sprintf("prop%d", i)
				names = append(names, fmt.Sprintf("%s: %s", propName, td.name))
				if i > 0 {
					encoding = append(encoding, ",")
				}
				encoding = append(encoding, fmt.Sprintf(`"%s"`, propName))
				encoding = append(encoding, ":")
				encoding = append(encoding, td.encoding...)
				propActions = append(propActions, PropertyAction{Name: propName, Action: td.action})
			}
			encoding = append(encoding, "}")
			value := AnyValue{Kind: ObjectValue, Object: propActions}
			objectTest := testDef{
				name:     "object(" + strings.Join(names, ", ") + ")",
				encoding: encoding,
				action:   f.valueTestFactory.Value(value, ""),
			}
			ret = append(ret, objectTest)
		}
	}
	return ret
}

func (f testFactory) makeValueListsOfLength(count int) []testDefs {
	if count == 0 {
		return []testDefs{testDefs{}}
	}
	previousLists := f.makeValueListsOfLength(count - 1)
	ret := []testDefs{}
	for _, previous := range previousLists {
		for _, elementTest := range f.makeScalarValueTests(false) {
			ret = append(ret, append(previous, elementTest))
		}
	}
	return ret
}
