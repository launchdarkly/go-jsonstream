package jwriter

// marshalInitialCapacity is the output buffer capacity that MarshalJSONWithWriter starts
// with, sized so that most documents marshal without reallocating.
const marshalInitialCapacity = 1000

// MarshalJSONWithWriter is a convenience method for implementing json.Marshaler to marshal to a
// byte slice with the default TokenWriter implementation.
func MarshalJSONWithWriter(writable Writable) ([]byte, error) {
	w := Writer{tw: newTokenWriterWithCapacity(marshalInitialCapacity)}
	writable.WriteToJSONWriter(&w)
	if err := w.Error(); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
