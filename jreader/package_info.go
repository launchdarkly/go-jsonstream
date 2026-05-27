// Package jreader provides an efficient mechanism for reading JSON data sequentially.
//
// The high-level API for this package, Writer, is designed to facilitate writing custom JSON
// marshaling logic concisely and reliably. Output is buffered in memory.
//
//	import (
//	    "gopkg.in/launchdarkly/jsonstream.v1/jreader"
//	)
//
//	type myStruct struct {
//	    value int
//	}
//
//	func (s *myStruct) ReadFromJSONReader(r *jreader.Reader) {
//	    // reading a JSON object structure like {"value":2}
//	    for obj := r.Object(); obj.Next; {
//	        if string(obj.Name()) == "value" {
//	            s.value = r.Int()
//	        }
//	    }
//	}
//
//	func ParseMyStructJSON() {
//	    var s myStruct
//	    r := jreader.NewReader([]byte(`{"value":2}`))
//	    s.ReadFromJSONReader(&r)
//	    fmt.Printf("%+v\n", s)
//	}
package jreader
