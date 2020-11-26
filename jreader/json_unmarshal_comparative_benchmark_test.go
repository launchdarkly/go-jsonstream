package jreader

import (
	"encoding/json"
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"
)

// These benchmarks perform equivalent actions to the ones in reader_benchmark_test.go, but using
// the default reflection-based mechanism from the json/encoding package, so we can see how much
// less efficient that is than our default implementation and the easyjson implementation.

func BenchmarkJSONUnmarshalComparatives(b *testing.B) {
	b.Run("Null", benchmarkReadNullJSONUnmarshal)
	b.Run("Boolean", benchmarkReadBooleanJSONUnmarshal)
	b.Run("NumberInt", benchmarkReadNumberIntJSONUnmarshal)
	b.Run("NumberFloat", benchmarkReadNumberFloatJSONUnmarshal)
	b.Run("String", benchmarkReadStringJSONUnmarshal)
	b.Run("ArrayOfBools", benchmarkReadArrayOfBoolsJSONUnmarshal)
	b.Run("ArrayOfStrings", benchmarkReadArrayOfStringsJSONUnmarshal)
	b.Run("Object", benchmarkReadObjectJSONUnmarshal)
}

func benchmarkReadNullJSONUnmarshal(b *testing.B) {
	data := []byte("null")
	var expected interface{} = nil
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val interface{}
		if err := json.Unmarshal(data, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != expected {
			b.FailNow()
		}
	}
}

func benchmarkReadBooleanJSONUnmarshal(b *testing.B) {
	data := []byte("true")
	expected := true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val bool
		if err := json.Unmarshal(data, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != expected {
			b.FailNow()
		}
	}
}

func benchmarkReadNumberIntJSONUnmarshal(b *testing.B) {
	data := []byte("1234")
	expected := 1234
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val int
		if err := json.Unmarshal(data, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != expected {
			b.FailNow()
		}
	}
}

func benchmarkReadNumberFloatJSONUnmarshal(b *testing.B) {
	data := []byte("1234.5")
	expected := 1234.5
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val float64
		if err := json.Unmarshal(data, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != expected {
			b.FailNow()
		}
	}
}

func benchmarkReadStringJSONUnmarshal(b *testing.B) {
	data := []byte(`"abc"`)
	expected := "abc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val string
		if err := json.Unmarshal(data, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != expected {
			b.FailNow()
		}
	}
}

func benchmarkReadArrayOfBoolsJSONUnmarshal(b *testing.B) {
	expected := commontest.MakeBools()
	data := commontest.MakeBoolsJSON(expected)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var vals []bool
		if err := json.Unmarshal(data, &vals); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if len(vals) < len(expected) {
			b.FailNow()
		}
	}
}

func benchmarkReadArrayOfStringsJSONUnmarshal(b *testing.B) {
	expected := commontest.MakeStrings()
	data := commontest.MakeStringsJSON(expected)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var vals []string
		if err := json.Unmarshal(data, &vals); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if len(vals) < len(expected) {
			b.FailNow()
		}
	}
}

func benchmarkReadObjectJSONUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var val ExampleStructWrapper
		if err := json.Unmarshal(commontest.ExampleStructData, &val); err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != ExampleStructWrapper(commontest.ExampleStructValue) {
			b.FailNow()
		}
	}
}
