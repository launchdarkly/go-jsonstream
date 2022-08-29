package jreader

import (
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/internal/commontest"
)

func BenchmarkReadNullNoAlloc(b *testing.B) {
	data := []byte("null")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		if err := r.Null(); err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkReadBooleanNoAlloc(b *testing.B) {
	data := []byte("true")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val := r.Bool()
		if !val || r.Error() != nil {
			b.FailNow()
		}
	}
}

func BenchmarkReadNumberIntNoAlloc(b *testing.B) {
	data := []byte("1234")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val := r.Int()
		failBenchmarkOnReaderError(b, &r)
		if val != 1234 {
			b.FailNow()
		}
	}
}

func BenchmarkReadNumberFloat(b *testing.B) {
	data := []byte("1234.5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val := r.Float64()
		failBenchmarkOnReaderError(b, &r)
		if val != 1234.5 {
			b.FailNow()
		}
	}
}

func BenchmarkReadString(b *testing.B) {
	data := []byte(`"abc"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val := r.String()
		failBenchmarkOnReaderError(b, &r)
		if val != "abc" {
			b.FailNow()
		}
	}
}

func BenchmarkReadArrayOfBools(b *testing.B) {
	expected := commontest.MakeBools()
	data := commontest.MakeBoolsJSON(expected)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var vals []bool
		r := NewReader(data)
		arr := r.Array()
		failBenchmarkOnReaderError(b, &r)
		for arr.Next() {
			val := r.Bool()
			failBenchmarkOnReaderError(b, &r)
			vals = append(vals, val)
		}
		if len(vals) < len(expected) {
			b.FailNow()
		}
	}
}

func BenchmarkReadArrayOfStrings(b *testing.B) {
	expected := commontest.MakeStrings()
	data := commontest.MakeStringsJSON(expected)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var vals []string
		r := NewReader(data)
		arr := r.Array()
		failBenchmarkOnReaderError(b, &r)
		for arr.Next() {
			val := r.String()
			failBenchmarkOnReaderError(b, &r)
			vals = append(vals, val)
		}
		if len(vals) < len(expected) {
			b.FailNow()
		}
	}
}

func BenchmarkReadArrayOfNullsNoAlloc(b *testing.B) {
	// This just verifies that simply parsing an array doesn't cause any allocations, if the values don't.
	data := []byte(`[null,null]`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		arr := r.Array()
		failBenchmarkOnReaderError(b, &r)
		if !arr.Next() {
			b.FailNow()
		}
		if err := r.Null(); err != nil {
			b.FailNow()
		}
		if !arr.Next() {
			b.FailNow()
		}
		failBenchmarkOnReaderError(b, &r)
		if arr.Next() {
			b.FailNow()
		}
	}
}

func BenchmarkReadObjectNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var val ExampleStructWrapper
		r := NewReader(commontest.ExampleStructData)
		val.ReadFromJSONReader(&r)
		failBenchmarkOnReaderError(b, &r)
		if val != ExampleStructWrapper(commontest.ExampleStructValue) {
			b.FailNow()
		}
	}
}

func BenchmarkReadArrayOfObjects(b *testing.B) {
	rawStructs := commontest.MakeStructs()
	data := commontest.MakeStructsJSON(rawStructs)
	var expected []ExampleStructWrapper
	for _, rawStruct := range rawStructs {
		expected = append(expected, ExampleStructWrapper(rawStruct))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values := make([]ExampleStructWrapper, 0)
		r := NewReader(data)
		for arr := r.Array(); arr.Next(); {
			var val ExampleStructWrapper
			val.ReadFromJSONReader(&r)
			values = append(values, val)
		}
		failBenchmarkOnReaderError(b, &r)
		for i, val := range values {
			if val != expected[i] {
				b.FailNow()
			}
		}
	}
}

func BenchmarkReadObjectWithRequiredPropsNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var val ExampleStructWrapperWithRequiredProps
		r := NewReader(commontest.ExampleStructData)
		val.ReadFromJSONReader(&r)
		failBenchmarkOnReaderError(b, &r)
		if val != ExampleStructWrapperWithRequiredProps(commontest.ExampleStructValue) {
			b.FailNow()
		}
	}
}

func failBenchmarkOnReaderError(b *testing.B, r *Reader) {
	if r.Error() != nil {
		b.Error(r.Error())
		b.FailNow()
	}
}
