package jwriter

import (
	"testing"

	"github.com/launchdarkly/go-jsonstream/v4/internal/commontest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSONWithWriter(t *testing.T) {
	data, err := MarshalJSONWithWriter(ExampleStructWrapper(commontest.ExampleStructValue))
	assert.NoError(t, err)
	assert.Equal(t, commontest.ExampleStructData, data)
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
