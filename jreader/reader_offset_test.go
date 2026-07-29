package jreader

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Offset is provided only by the default build, like RawValue; its tests carry the same tag.

const jsonWhitespace = " \t\r\n"

func TestOffsetCapturesSpanAroundInPlaceParse(t *testing.T) {
	input := []byte(`{"first":1,"obj":{"x":[1,2],"y":"z"},"last":2}`)
	r := NewReader(input)
	var span []byte
	for obj := r.Object(); obj.Next(); {
		if string(obj.Name()) != "obj" {
			require.NoError(t, r.SkipValue())
			continue
		}
		start := r.Offset()
		require.NoError(t, r.SkipValue()) // stands in for any in-place parse of the value
		span = input[start:r.Offset()]
	}
	require.NoError(t, r.Error())
	assert.Equal(t, `{"x":[1,2],"y":"z"}`, string(bytes.Trim(span, jsonWhitespace)))
}

func TestOffsetSpanMayIncludeSurroundingWhitespace(t *testing.T) {
	input := []byte(`{ "obj" : { "x" : 1 } , "z" : 2 }`)
	r := NewReader(input)
	var span []byte
	for obj := r.Object(); obj.Next(); {
		if string(obj.Name()) != "obj" {
			require.NoError(t, r.SkipValue())
			continue
		}
		start := r.Offset()
		require.NoError(t, r.SkipValue())
		span = input[start:r.Offset()]
	}
	require.NoError(t, r.Error())
	// Whitespace around the value may be captured (and is trimmed by the caller); whitespace
	// inside the value is preserved verbatim.
	assert.Equal(t, `{ "x" : 1 }`, string(bytes.Trim(span, jsonWhitespace)))
}

func TestOffsetAtStartAndAfterWholeValue(t *testing.T) {
	input := []byte(`  [1,2,3]`)
	r := NewReader(input)
	assert.Equal(t, 0, r.Offset())
	require.NoError(t, r.SkipValue())
	assert.Equal(t, len(input), r.Offset())
}

func TestTokenReaderOffsetWithUnreadToken(t *testing.T) {
	// A failed Null probe pushes the token back; Offset must report the pushed-back token's
	// starting position, not the position after it.
	tr := newTokenReader([]byte(`  true`))
	isNull, err := tr.Null()
	require.NoError(t, err)
	require.False(t, isNull)
	assert.Equal(t, 2, tr.Offset())
	raw, err := tr.RawValue()
	require.NoError(t, err)
	assert.Equal(t, `true`, string(raw))
	assert.Equal(t, 6, tr.Offset())
}

func TestOffsetSpanIsValidJSONForEveryValueType(t *testing.T) {
	input := []byte(`{"null":null,"bool":true,"num":-1.5,"str":"s","arr":[[]],"obj":{"a":{}}}`)
	expected := map[string]string{
		"null": `null`, "bool": `true`, "num": `-1.5`, "str": `"s"`, "arr": `[[]]`, "obj": `{"a":{}}`,
	}
	r := NewReader(input)
	got := map[string]string{}
	for obj := r.Object(); obj.Next(); {
		name := string(obj.Name())
		start := r.Offset()
		require.NoError(t, r.SkipValue())
		got[name] = string(bytes.Trim(input[start:r.Offset()], jsonWhitespace))
	}
	require.NoError(t, r.Error())
	assert.Equal(t, expected, got)
}
