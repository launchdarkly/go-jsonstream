package jwriter

import (
	"encoding/json"
	"fmt"
)

func ExampleObjectState_Property() {
	myCustomMarshaler := func(w *Writer) {
		subObject := w.Object()
		subObject.Bool("yes", true)
		subObject.End()
	}

	w := NewWriter()

	obj := w.Object()
	myCustomMarshaler(obj.Property("subObject"))
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"subObject":{"yes":true}}
}

func ExampleObjectState_Null() {
	w := NewWriter()
	obj := w.Object()
	obj.Null("property")
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":null}
}

func ExampleObjectState_Bool() {
	w := NewWriter()
	obj := w.Object()
	obj.Bool("property", true)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":true}
}

func ExampleObjectState_Int() {
	w := NewWriter()
	obj := w.Object()
	obj.Int("property", 123)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":123}
}

func ExampleObjectState_Float64() {
	w := NewWriter()
	obj := w.Object()
	obj.Float64("property", 1234.5)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":1234.5}
}

func ExampleObjectState_String() {
	w := NewWriter()
	obj := w.Object()
	obj.String("property", `string says "hello"`)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":"string says \"hello\""}
}

func ExampleObjectState_Array() {
	w := NewWriter()
	obj := w.Object()
	arr := obj.Array("property")
	arr.Int(1)
	arr.Int(2)
	arr.End()
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":[1,2]}
}

func ExampleObjectState_Object() {
	w := NewWriter()
	obj := w.Object()
	subObj := obj.Object("property")
	subObj.Int("value", 1)
	subObj.End()
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"property":{"value":1}}
}

func ExampleObjectState_OptBool() {
	w := NewWriter()
	obj := w.Object()
	obj.OptBool("notPresent", false, true)
	obj.OptBool("present", true, true)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"present":true}
}

func ExampleObjectState_OptInt() {
	w := NewWriter()
	obj := w.Object()
	obj.OptInt("notPresent", false, 123)
	obj.OptInt("present", true, 456)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"present":456}
}

func ExampleObjectState_OptFloat64() {
	w := NewWriter()
	obj := w.Object()
	obj.OptFloat64("property", false, 1234.5)
	obj.OptFloat64("present", true, 6)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"present":6}
}

func ExampleObjectState_OptString() {
	w := NewWriter()
	obj := w.Object()
	obj.OptString("notPresent", false, "a")
	obj.OptString("present", true, "b")
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"present":"b"}
}

func ExampleObjectState_Raw() {
	data := json.RawMessage(`{"value":1}`)
	w := NewWriter()
	obj := w.Object()
	obj.Raw("data", data)
	obj.End()

	fmt.Println(string(w.Bytes()))
	// Output: {"data":{"value":1}}
}
