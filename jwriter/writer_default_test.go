//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jwriter

// These tests pin buffer-management behaviors of the default token writer (growth
// amortization and the pre-sized marshal buffer), so they apply only to the default
// implementation.

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriterReallocationsAreAmortized(t *testing.T) {
	// Growing the output buffer must at least double its capacity when tokens are
	// written, so the number of reallocations stays logarithmic in the output size.
	// 5000 tokens of each kind produce far under 64 KiB of output, so a doubling ladder
	// starting at 64 bytes can take at most ~11 steps; the bound below fails if a token
	// path falls back to append's ~1.25x growth. (The bound has slack because delimiter
	// writes do not reserve; see the streamableBuffer comment.)
	const maxSteps = 14
	writeToken := map[string]func(arr *ArrayState, i int){
		"string": func(arr *ArrayState, i int) { arr.String("value\twith\n\"escaped chars\"") },
		"int":    func(arr *ArrayState, i int) { arr.Int(123456789 + i) },
		"bool":   func(arr *ArrayState, i int) { arr.Bool(i%2 == 0) },
		"raw":    func(arr *ArrayState, i int) { arr.Raw(json.RawMessage(`{"k":[1,2,3]}`)) },
	}
	for kind, write := range writeToken {
		t.Run(kind, func(t *testing.T) {
			w := NewWriter()
			arr := w.Array()
			steps := 0
			lastCap := -1
			for i := 0; i < 5000; i++ {
				write(&arr, i)
				if c := cap(w.tw.buf.Bytes()); c != lastCap {
					steps++
					lastCap = c
				}
			}
			arr.End()
			require.NoError(t, w.Error())
			require.LessOrEqual(t, steps, maxSteps)
		})
	}
}

// marshalAllocationTestWritable emits a document of several hundred bytes: larger than the
// 64-byte buffer a plain NewWriter starts with, but within MarshalJSONWithWriter's initial
// capacity, so the allocation assertion below fails if that capacity is ever lost.
type marshalAllocationTestWritable struct{}

func (marshalAllocationTestWritable) WriteToJSONWriter(w *Writer) {
	arr := w.Array()
	for i := 0; i < 20; i++ {
		arr.String("string value for the allocation test")
	}
	arr.End()
}

func TestMarshalJSONWithWriterAllocations(t *testing.T) {
	var writable Writable = marshalAllocationTestWritable{}
	allocs := testing.AllocsPerRun(500, func() {
		data, err := MarshalJSONWithWriter(writable)
		if err != nil || len(data) == 0 {
			t.Fail()
		}
	})
	// One allocation for the Writer (it escapes through the Writable interface) and one
	// for the output buffer, allocated once at full size.
	require.LessOrEqual(t, allocs, 2.0)
}
