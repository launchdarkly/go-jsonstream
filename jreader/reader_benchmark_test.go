package jreader

import (
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"
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
		val, _, err := r.Bool(false)
		if !val || err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkReadNumberIntNoAlloc(b *testing.B) {
	data := []byte("1234")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val, _, err := r.Int(false)
		if val != 1234 || err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkReadNumberFloat(b *testing.B) {
	data := []byte("1234.5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val, _, err := r.Float64(false)
		if val != 1234.5 || err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkReadString(b *testing.B) {
	data := []byte(`"abc"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewReader(data)
		val, _, err := r.String(false)
		if val != "abc" || err != nil {
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
		arr, err := r.Array(false)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		for arr.Next() {
			val, _, err := r.Bool(false)
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
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
		arr, err := r.Array(false)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		for arr.Next() {
			val, _, err := r.String(false)
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
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
		arr, err := r.Array(false)
		if err != nil {
			b.FailNow()
		}
		if !arr.Next() {
			b.FailNow()
		}
		if err := r.Null(); err != nil {
			b.FailNow()
		}
		if !arr.Next() {
			b.FailNow()
		}
		if err := r.Null(); err != nil {
			b.FailNow()
		}
		if arr.Next() {
			b.FailNow()
		}
	}
}

func BenchmarkReadObjectNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var val ExampleStructWrapper
		r := NewReader(commontest.ExampleStructData)
		err := val.ReadFromJSONReader(&r)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if val != ExampleStructWrapper(commontest.ExampleStructValue) {
			b.FailNow()
		}
	}
}
