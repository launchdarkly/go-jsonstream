package jwriter

import (
	"bytes"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v2/internal/commontest"
)

func BenchmarkWriteNull(b *testing.B) {
	expected := []byte("null")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		w.Null()
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteBoolean(b *testing.B) {
	expected := []byte("true")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		w.Bool(true)
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteNumberInt(b *testing.B) {
	expected := []byte("123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		w.Int(123)
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteNumberFloat(b *testing.B) {
	expected := []byte("1234.5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		w.Float64(1234.5)
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteString(b *testing.B) {
	expected := []byte(`"abc"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		w.String("abc")
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteArrayOfBools(b *testing.B) {
	vals := commontest.MakeBools()
	expected := commontest.MakeBoolsJSON(vals)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		arr := w.Array()
		for _, val := range vals {
			arr.Bool(val)
		}
		arr.End()
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteArrayOfStrings(b *testing.B) {
	vals := commontest.MakeStrings()
	expected := commontest.MakeStringsJSON(vals)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		arr := w.Array()
		for _, val := range vals {
			arr.String(val)
		}
		arr.End()
		benchmarkExpectWriterOutput(b, &w, expected)
	}
}

func BenchmarkWriteObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := NewWriter()
		ExampleStructWrapper(commontest.ExampleStructValue).WriteToJSONWriter(&w)
		benchmarkExpectWriterOutput(b, &w, commontest.ExampleStructData)
	}
}

func benchmarkExpectWriterOutput(b *testing.B, w *Writer, expectedJSON []byte) {
	if err := w.Error(); err != nil {
		b.Error(err)
		b.FailNow()
	}
	if !bytes.Equal(expectedJSON, w.Bytes()) {
		b.FailNow()
	}
}

func BenchmarkWriteObjectToNoOpWriterNoAllocs(b *testing.B) {
	// The purpose of this benchmark is to ensure that nothing is escaping to the heap simply
	// as result of calling the Name or Maybe methods (as it might if we hadn't been
	// careful about our use of pointers). We're preinitializing the Writer to already have an
	// error, so it won't produce any output.
	w := NewWriter()
	obj := w.Object()
	w.AddError(noOpWriterError{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Name("prop1").Int(1)
		obj.Maybe("prop2", true).Int(2)
		obj.Maybe("prop3", false).Int(3)
	}
}

func BenchmarkStreamingWriterArrayOfStrings(b *testing.B) {
	vals := commontest.MakeStrings()
	expected := commontest.MakeStringsJSON(vals)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		w := NewStreamingWriter(&buf, 50)
		arr := w.Array()
		for _, val := range vals {
			arr.String(val)
		}
		arr.End()
		w.Flush()
		output := buf.Bytes()
		if !bytes.Equal(expected, output) {
			b.FailNow()
		}
	}
}
