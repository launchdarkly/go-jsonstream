// Package jwriter provides an efficient mechanism for writing JSON data sequentially.
//
// The high-level API for this package, Writer, is designed to facilitate writing custom JSON
// marshaling logic concisely and reliably. Output is buffered in memory, and can optionally be
// dumped to an io.Writer at intervals.
//
// The underlying low-level token writing mechanism has two available implementations. The default
// implementation has no external dependencies. For interoperability with the easyjson library
// (https://github.com/mailru/easyjson), there is also an implementation that delegates to the
// easyjson streaming writer; this is enabled by setting the build tag "launchdarkly_easyjson".
// Be aware that by default, easyjson uses Go's "unsafe" package (https://pkg.go.dev/unsafe),
// which may not be available on all platforms.
package jwriter
