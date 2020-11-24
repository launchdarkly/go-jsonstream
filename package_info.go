// Package jsonstream provides a fast streaming JSON encoding and decoding mechanism.
//
// In the default implementation, this package has no external dependencies. Setting the build
// tag "launchdarkly_easyjson" causes it to use https://github.com/mailru/easyjson as its
// underlying reader/writer implementation.
//
// For more information, see: https://github.com/launchdarkly/go-jsonstream
package jsonstream
