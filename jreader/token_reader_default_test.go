//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

// isEasyJSON is used in tests to e.g. expect different allocation behavior depending
// on which backend is in use.
const isEasyJSON = false
