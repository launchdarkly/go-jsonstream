package jwriter

// This function tells the writer tests that we shouldn't expect to see hex escape sequences in the output.
func tokenWriterWillEncodeAsHex(_ rune) bool { //nolint:unused // linter is confused
	return false
}
