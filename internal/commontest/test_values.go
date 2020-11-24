package commontest

import (
	"fmt"
	"strconv"
)

// This file is used by test_factory.go to define a standard set of JSON scalar test values.

type encodingBehavior struct {
	encodeAsHex func(rune) bool
	forParsing  bool
}

type testValue struct {
	name     string
	encoding string
	value    AnyValue
}

type numberTestValueBase struct {
	name             string
	val              float64
	encoding         string
	simplestEncoding string
}

type stringTestValueBase struct {
	name     string
	val      string
	encoding string
}

func makeBoolTestValues() []testValue {
	return []testValue{
		{"bool true", "true", AnyValue{Kind: BoolValue, Bool: true}},
		{"bool false", "false", AnyValue{Kind: BoolValue, Bool: false}},
	}
}

func makeNumberTestValues(encodingBehavior encodingBehavior) []testValue {
	var ret []testValue
	for _, v := range []numberTestValueBase{
		{"zero", 0, "0", ""},
		{"int", 3, "3", ""},
		{"int negative", -3, "-3", ""},
		{"int large", 1603312301195, "1603312301195", ""}, // enough magnitude for a millisecond timestamp
		{"float", 3.5, "3.5", ""},
		{"float negative", -3.5, "-3.5", ""},
		{"float with exp", 3500, "3.5e3", "3500"},
		{"float with Exp", 3500, "3.5E3", "3500"},
		{"float with exp+", 3500, "3.5e+3", "3500"},
		{"float with exp-", 0.0035, "3.5e-3", "0.0035"},
	} {
		enc := v.encoding
		if !encodingBehavior.forParsing && v.simplestEncoding != "" {
			enc = v.simplestEncoding
		}
		ret = append(ret, testValue{"number " + v.name, enc, AnyValue{Kind: NumberValue, Number: v.val}})
	}
	return ret
}

func makeStringTestValues(encodingBehavior encodingBehavior, allPermutations bool) []testValue {
	base := []stringTestValueBase{
		{name: "empty", val: "", encoding: `""`},
		{name: "simple", val: "abc", encoding: `"abc"`},
	}
	allEscapeTests := []stringTestValueBase{}
	if allPermutations {
		baseEscapeTests := []stringTestValueBase{
			{val: `"`, encoding: `\"`},
			{val: `\`, encoding: `\\`},
			{val: "\x05", encoding: `\u0005`},
			{val: "\x1c", encoding: `\u001c`},
			{val: "🦜🦄😂🧶😻 yes", encoding: "🦜🦄😂🧶😻 yes"}, // unescaped multi-byte characters are allowed
		}
		addControlChar := func(str, shortEncoding string) {
			hex := strconv.FormatInt(int64(str[0]), 16)
			if len(hex) == 1 {
				hex = "0" + hex
			}
			encodeAsHex := false
			if encodingBehavior.encodeAsHex != nil {
				encodeAsHex = encodingBehavior.encodeAsHex(rune(str[0]))
			}
			if !encodingBehavior.forParsing && !encodeAsHex {
				baseEscapeTests = append(baseEscapeTests, stringTestValueBase{val: str, encoding: shortEncoding})
			}
			if encodingBehavior.forParsing || encodeAsHex {
				baseEscapeTests = append(baseEscapeTests, stringTestValueBase{val: str, encoding: `\u00` + hex})
			}
		}
		addControlChar("\b", `\b`)
		addControlChar("\t", `\t`)
		addControlChar("\n", `\n`)
		addControlChar("\f", `\f`)
		addControlChar("\r", `\r`)
		if encodingBehavior.forParsing {
			// These escapes are not used when writing, but may be encountered when parsing
			baseEscapeTests = append(baseEscapeTests, stringTestValueBase{val: "/", encoding: `\/`})
			baseEscapeTests = append(baseEscapeTests, stringTestValueBase{val: "も", encoding: `\u3082`})
		}
		for _, et := range baseEscapeTests {
			allEscapeTests = append(allEscapeTests, et)
			addTransformed := func(valFn func(string) string, encFn func(string) string) {
				tt := stringTestValueBase{
					val:      valFn(et.val),
					encoding: encFn(et.encoding),
				}
				allEscapeTests = append(allEscapeTests, tt)
			}
			for _, f := range []string{"%sabcd", "abcd%s", "ab%scd"} {
				addTransformed(func(s string) string { return fmt.Sprintf(f, s) },
					func(s string) string { return fmt.Sprintf(f, s) })
			}
			for _, et2 := range baseEscapeTests {
				for _, f := range []string{"%s%sabcd", "ab%s%scd", "a%sbc%sd", "abcd%s%s"} {
					addTransformed(func(s string) string { return fmt.Sprintf(f, s, et2.val) },
						func(s string) string { return fmt.Sprintf(f, s, et2.encoding) })
				}
			}
		}
	} else {
		// When we're testing nested data structures, we don't need to cover all those permutations of
		// escape sequences-- we can assume that the same string encoding logic applies as it would for
		// a single value.
		allEscapeTests = []stringTestValueBase{{val: "simple\tescape", encoding: `simple\tescape`}}
	}
	for i, et := range allEscapeTests {
		st := stringTestValueBase{
			name:     fmt.Sprintf("with escapes %d", i+1),
			val:      et.val,
			encoding: `"` + et.encoding + `"`,
		}
		base = append(base, st)
	}
	ret := make([]testValue, 0, len(base))
	for _, b := range base {
		ret = append(ret, testValue{name: "string " + b.name, encoding: b.encoding,
			value: AnyValue{Kind: StringValue, String: b.val}})
	}
	return ret
}

type boolTestValue struct {
	name     string
	val      bool
	encoding string
}

type numberTestValue struct {
	name             string
	val              float64
	encoding         string
	simplestEncoding string
}

type stringTestValue struct {
	name     string
	val      string
	encoding string
}

func makeValidBoolTestValues() []boolTestValue {
	return []boolTestValue{
		{"true", true, "true"},
		{"false", false, "false"},
	}
}

func makeValidNumberTestValues() []numberTestValue {
	return []numberTestValue{
		{"zero", 0, "0", ""},
		{"int", 3, "3", ""},
		{"int negative", -3, "-3", ""},
		{"int large", 1603312301195, "1603312301195", ""}, // enough magnitude for a millisecond timestamp
		{"float", 3.5, "3.5", ""},
		{"float negative", -3.5, "-3.5", ""},
		{"float with exp", 3500, "3.5e3", "3500"},
		{"float with Exp", 3500, "3.5E3", "3500"},
		{"float with exp+", 3500, "3.5e+3", "3500"},
		{"float with exp-", 0.0035, "3.5e-3", "0.0035"},
	}
}

func makeValidStringTestValues(encodingBehavior encodingBehavior, allPermutations bool) []stringTestValue {
	ret := []stringTestValue{
		{name: "empty", val: "", encoding: `""`},
		{name: "simple", val: "abc", encoding: `"abc"`},
	}
	allEscapeTests := []stringTestValue{}
	if allPermutations {
		baseEscapeTests := []stringTestValue{
			{val: `"`, encoding: `\"`},
			{val: `\`, encoding: `\\`},
			{val: "\x05", encoding: `\u0005`},
			{val: "\x1c", encoding: `\u001c`},
			{val: "🦜🦄😂🧶😻 yes", encoding: "🦜🦄😂🧶😻 yes"}, // unescaped multi-byte characters are allowed
		}
		addControlChar := func(str, shortEncoding string) {
			hex := strconv.FormatInt(int64(str[0]), 16)
			if len(hex) == 1 {
				hex = "0" + hex
			}
			encodeAsHex := false
			if encodingBehavior.encodeAsHex != nil {
				encodeAsHex = encodingBehavior.encodeAsHex(rune(str[0]))
			}
			if !encodingBehavior.forParsing && !encodeAsHex {
				baseEscapeTests = append(baseEscapeTests, stringTestValue{val: str, encoding: shortEncoding})
			}
			if encodingBehavior.forParsing || encodeAsHex {
				baseEscapeTests = append(baseEscapeTests, stringTestValue{val: str, encoding: `\u00` + hex})
			}
		}
		addControlChar("\b", `\b`)
		addControlChar("\t", `\t`)
		addControlChar("\n", `\n`)
		addControlChar("\f", `\f`)
		addControlChar("\r", `\r`)
		if encodingBehavior.forParsing {
			// These escapes are not used when writing, but may be encountered when parsing
			baseEscapeTests = append(baseEscapeTests, stringTestValue{val: "/", encoding: `\/`})
			baseEscapeTests = append(baseEscapeTests, stringTestValue{val: "も", encoding: `\u3082`})
		}
		for _, et := range baseEscapeTests {
			allEscapeTests = append(allEscapeTests, et)
			addTransformed := func(valFn func(string) string, encFn func(string) string) {
				tt := stringTestValue{
					val:      valFn(et.val),
					encoding: encFn(et.encoding),
				}
				allEscapeTests = append(allEscapeTests, tt)
			}
			for _, f := range []string{"%sabcd", "abcd%s", "ab%scd"} {
				addTransformed(func(s string) string { return fmt.Sprintf(f, s) },
					func(s string) string { return fmt.Sprintf(f, s) })
			}
			for _, et2 := range baseEscapeTests {
				for _, f := range []string{"%s%sabcd", "ab%s%scd", "a%sbc%sd", "abcd%s%s"} {
					addTransformed(func(s string) string { return fmt.Sprintf(f, s, et2.val) },
						func(s string) string { return fmt.Sprintf(f, s, et2.encoding) })
				}
			}
		}
	} else {
		// When we're testing nested data structures, we don't need to cover all those permutations of
		// escape sequences-- we can assume that the same string encoding logic applies as it would for
		// a single value.
		allEscapeTests = []stringTestValue{{val: "simple\tescape", encoding: `simple\tescape`}}
	}
	for i, et := range allEscapeTests {
		st := stringTestValue{
			name:     fmt.Sprintf("with escapes %d", i+1),
			val:      et.val,
			encoding: `"` + et.encoding + `"`,
		}
		ret = append(ret, st)
	}
	return ret
}

func MakeValidDelimiters() []byte {
	return []byte{',', ':', '{', '[', '}', ']'}
}

func MakeWhitespaceOptions() map[string]string {
	return map[string]string{"spaces": "  ", "tab": "\t", "newline": "\n"}
}
