package jwriter

import (
	"fmt"
)

func ExampleObjectState_Name() {
	myCustomMarshaler := func(w *Writer) {
		subObject := w.Object()
		subObject.Name("yes").Bool(true)
		subObject.End()
	}

	w := NewWriter()

	obj := w.Object()
	myCustomMarshaler(obj.Name("subObject"))
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"subObject":{"yes":true}}
}

func ExampleObjectState_Maybe() {
	w := NewWriter()
	obj := w.Object()
	obj.Maybe("notPresent", false).Int(1)
	obj.Maybe("present", true).Int(2)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"present":2}
}
