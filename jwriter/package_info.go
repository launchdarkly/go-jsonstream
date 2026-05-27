// Package jwriter provides an efficient mechanism for writing JSON data sequentially.
//
// The high-level API for this package, Writer, is designed to facilitate writing custom JSON
// marshaling logic concisely and reliably. Output is buffered in memory.
//
//	import (
//	    "gopkg.in/launchdarkly/jsonstream.v1/jwriter"
//	)
//
//	type myStruct struct {
//	    value int
//	}
//
//	func (s myStruct) WriteToJSONWriter(w *jwriter.Writer) {
//	    obj := w.Object() // writing a JSON object structure like {"value":2}
//	    obj.Property("value").Int(s.value)
//	    obj.End()
//	}
//
//	func PrintMyStructJSON(s myStruct) {
//	    w := jwriter.NewWriter()
//	    s.WriteToJSONWriter(&w)
//	    fmt.Println(string(w.Bytes())
//	}
//
// Output can optionally be dumped to an io.Writer at intervals to avoid allocating a large buffer:
//
//	func WriteToHTTPResponse(s myStruct, resp http.ResponseWriter) {
//	    resp.Header.Add("Content-Type", "application/json")
//	    w := jwriter.NewStreamingWriter(resp, 1000)
//	    myStruct.WriteToJSONWriter(&w)
//	}
package jwriter
