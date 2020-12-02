// Package jsonstream provides a fast streaming JSON encoding and decoding mechanism.
//
// The base package is empty; see the jreader and jwriter subpackages.
//
// In the default implementation, these packages have no external dependencies. Setting the build
// tag "launchdarkly_easyjson" causes them to use https://github.com/mailru/easyjson as the
// underlying reader/writer implementation.
//
// For more information, see: https://github.com/launchdarkly/go-jsonstream
package jsonstream
