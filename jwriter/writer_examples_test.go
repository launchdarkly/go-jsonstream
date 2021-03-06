package jwriter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func ExampleNewWriter() {
	w := NewWriter()
	obj := w.Object()
	obj.Name("property").String("value")
	obj.End()
	fmt.Println(string(w.Bytes()))
	// Output: {"property":"value"}
}

func ExampleNewStreamingWriter() {
	w := NewStreamingWriter(os.Stdout, 10)
	obj := w.Object()
	obj.Name("property").String("value")
	obj.End()
	w.Flush()
	// Output: {"property":"value"}
}

func ExampleWriter_AddError() {
	w := NewWriter()
	obj := w.Object()
	obj.Name("prop1").Bool(true)
	w.AddError(errors.New("sorry, we can't serialize this after all"))
	obj.Name("prop2").Bool(true) // no output is generated here because the Writer has already failed
	fmt.Println("error is:", w.Error())
	fmt.Println("buffer is:", string(w.Bytes()))
	// Output: error is: sorry, we can't serialize this after all
	// buffer is: {"prop1":true
}

func ExampleWriter_Null() {
	w := NewWriter()
	w.Null()
	fmt.Println(string(w.Bytes()))
	// Output: null
}

func ExampleWriter_Bool() {
	w := NewWriter()
	w.Bool(true)
	fmt.Println(string(w.Bytes()))
	// Output: true
}

func ExampleWriter_BoolOrNull() {
	w := NewWriter()
	w.BoolOrNull(false, true)
	fmt.Println(string(w.Bytes()))
	// Output: null
}

func ExampleWriter_Int() {
	w := NewWriter()
	w.Int(123)
	fmt.Println(string(w.Bytes()))
	// Output: 123
}

func ExampleWriter_IntOrNull() {
	w := NewWriter()
	w.IntOrNull(false, 1)
	fmt.Println(string(w.Bytes()))
	// Output: null
}

func ExampleWriter_Float64() {
	w := NewWriter()
	w.Float64(1234.5)
	fmt.Println(string(w.Bytes()))
	// Output: 1234.5
}

func ExampleWriter_Float64OrNull() {
	w := NewWriter()
	w.Float64OrNull(false, 1)
	fmt.Println(string(w.Bytes()))
	// Output: null
}

func ExampleWriter_String() {
	w := NewWriter()
	w.String(`string says "hello"`)
	fmt.Println(string(w.Bytes()))
	// Output: "string says \"hello\""
}

func ExampleWriter_StringOrNull() {
	w := NewWriter()
	w.StringOrNull(false, "no")
	fmt.Println(string(w.Bytes()))
	// Output: null
}

func ExampleWriter_Array() {
	w := NewWriter()
	arr := w.Array()
	arr.Bool(true)
	arr.Int(3)
	arr.End()
	fmt.Println(string(w.Bytes()))
	// Output: [true,3]
}

func ExampleWriter_Object() {
	w := NewWriter()
	obj := w.Object()
	obj.Name("boolProperty").Bool(true)
	obj.Name("intProperty").Int(3)
	obj.End()
	fmt.Println(string(w.Bytes()))
	// Output: {"boolProperty":true,"intProperty":3}
}

func ExampleWriter_Raw() {
	data := json.RawMessage(`{"value":1}`)
	w := NewWriter()
	w.Raw(data)

	fmt.Println(string(w.Bytes()))
	// Output: {"value":1}
}
