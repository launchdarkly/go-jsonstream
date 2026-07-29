package jreader

// String attempts to read a string value.
//
// If there is a parsing error, or the next value is not a string, the return value is "" and
// the Reader enters a failed state, which you can detect with Error(). Types other than string
// are never converted to strings.
func (r *Reader) String() string {
	return string(r.StringAsBytes())
}

// StringAsBytes attempts to read a string value, returning a byte slice that indexes into the
// original JSON bytes.  This method can be used instead of String to avoid garbage creation,
// but care must be taken to avoid modifying the returned byte slice.
//
// If there is a parsing error, or the next value is not a string, the return value is nil and
// the Reader enters a failed state, which you can detect with Error(). Types other than string
// are never converted to strings.
func (r *Reader) StringAsBytes() []byte {
	r.awaitingReadValue = false
	if r.err != nil {
		return nil
	}
	val, err := r.tr.StringAsBytes()
	if err != nil {
		r.err = err
		return nil
	}
	return val
}

// RawValue consumes the next JSON value of any type and returns its raw bytes verbatim. For an
// array or object value, this includes everything up to and including the matching close
// delimiter. The returned slice references the Reader's input data and is valid only as long as
// the input is; it must not be modified.
//
// The value is fully validated, so malformed JSON produces a parsing error rather than being
// returned. Scalars are validated by the tokenizer, whose grammar is RFC 8259 compliant. For an
// array or object, the byte boundary is located by scanning to the matching close delimiter and
// the captured bytes are then checked with encoding/json, since the boundary scan does not
// interpret the content in between.
//
// If there is a parsing error, or there is no next value, the return value is nil and the Reader
// enters a failed state, which you can detect with Error().
func (r *Reader) RawValue() []byte {
	r.awaitingReadValue = false
	if r.err != nil {
		return nil
	}
	val, err := r.tr.RawValue()
	if err != nil {
		r.err = err
		return nil
	}
	return val
}

// Offset returns the Reader's current byte offset within its original input: the position of
// the next byte that will be consumed, or, if a token has been read ahead and pushed back, the
// position where that token starts.
//
// Its primary use is capturing the raw bytes of a value that is parsed in place, without the
// separate boundary scan and validation pass that RawValue performs: record Offset before and
// after parsing the value, then slice the original input. Since the tokenizer fully validates
// everything it parses, the captured span is known-valid JSON. The offset may point at
// whitespace preceding the next token (and, after a value, at whitespace that was skipped before
// a following token was read ahead), so callers capturing a span should trim JSON whitespace
// (space, tab, carriage return, newline) from both ends.
func (r *Reader) Offset() int {
	return r.tr.Offset()
}
