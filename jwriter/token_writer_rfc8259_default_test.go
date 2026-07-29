package jwriter

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NaN and infinities cannot be represented in JSON, so the writer must report an error rather
// than emit invalid output.
func TestWriterRejectsNonFiniteFloat64(t *testing.T) {
	for _, value := range []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	} {
		t.Run("", func(t *testing.T) {
			w := NewWriter()
			w.Float64(value)
			require.Error(t, w.Error())
		})
	}
}

func TestWriterAcceptsFiniteFloat64(t *testing.T) {
	w := NewWriter()
	w.Float64(1.5)
	require.NoError(t, w.Error())
	assert.Equal(t, "1.5", string(w.Bytes()))
}
