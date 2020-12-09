package jwriter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/launchdarkly/go-jsonstream/internal/commontest"
)

// These benchmarks perform equivalent actions to the ones in writer_benchmark_test.go, but using
// the default reflection-based mechanism from the json/encoding package, so we can see how much
// less efficient that is than our default implementation and the easyjson implementation.

func BenchmarkJSONMarshalComparatives(b *testing.B) {
	b.Run("Null", benchmarkWriteNullJSONMarshal)
	b.Run("Boolean", benchmarkWriteBooleanJSONMarshal)
	b.Run("NumberInt", benchmarkWriteNumberIntJSONMarshal)
	b.Run("NumberFloat", benchmarkWriteNumberFloatJSONMarshal)
	b.Run("String", benchmarkWriteStringJSONMarshal)
	b.Run("ArrayOfBools", benchmarkWriteArrayOfBoolsJSONMarshal)
	b.Run("ArrayOfStrings", benchmarkWriteArrayOfStringsJSONMarshal)
	b.Run("Object", benchmarkWriteObjectJSONMarshal)
}

func benchmarkWriteNullJSONMarshal(b *testing.B) {
	expected := []byte("null")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(nil)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteBooleanJSONMarshal(b *testing.B) {
	expected := []byte("true")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(true)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteNumberIntJSONMarshal(b *testing.B) {
	expected := []byte("123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(123)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteNumberFloatJSONMarshal(b *testing.B) {
	expected := []byte("1234.5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(1234.5)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteStringJSONMarshal(b *testing.B) {
	expected := []byte(`"abc"`)
	val := "abc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(val)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteArrayOfBoolsJSONMarshal(b *testing.B) {
	vals := commontest.MakeBools()
	expected := commontest.MakeBoolsJSON(vals)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(vals)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteArrayOfStringsJSONMarshal(b *testing.B) {
	vals := commontest.MakeStrings()
	expected := commontest.MakeStringsJSON(vals)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(vals)
		benchmarkExpectJSONMarshalOutput(b, err, data, expected)
	}
}

func benchmarkWriteObjectJSONMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(ExampleStructWrapper(commontest.ExampleStructValue))
		benchmarkExpectJSONMarshalOutput(b, err, data, commontest.ExampleStructData)
	}
}

func benchmarkExpectJSONMarshalOutput(b *testing.B, err error, actualJSON []byte, expectedJSON []byte) {
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	if !bytes.Equal(expectedJSON, actualJSON) {
		b.FailNow()
	}
}
