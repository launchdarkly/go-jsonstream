// Package commontest provides test logic for JSON reading and writing.
//
// To ensure that the go-jsonstream types perform correctly with a wide range of inputs and outputs,
// we generate many permutations (single scalar values of various types; numbers in different formats;
// strings with or without escape characters at different positions; arrays and objects with different
// numbers of elements/properties) which are tested for both readers and writers. For readers, we also
// test various permutations of invalid input.
//
// Reader and writer tests are run against the high-level APIs (Reader, Writer) and the default
// implementations of the low-level APIs (tokenReader, tokenWriter).
package commontest
