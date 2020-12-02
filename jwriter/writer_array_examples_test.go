package jwriter

import (
	"encoding/json"
	"fmt"
)

func ExampleArrayState_Null() {
	w := NewWriter()
	arr := w.Array()
	arr.Null()
	arr.Null()
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [null,null]
}

func ExampleArrayState_Bool() {
	w := NewWriter()
	arr := w.Array()
	arr.Bool(true)
	arr.Bool(false)
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [true,false]
}

func ExampleArrayState_Int() {
	w := NewWriter()
	arr := w.Array()
	arr.Int(123)
	arr.Int(456)
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [123,456]
}

func ExampleArrayState_Float64() {
	w := NewWriter()
	arr := w.Array()
	arr.Float64(1234.5)
	arr.Float64(6)
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [1234.5,6]
}

func ExampleArrayState_String() {
	w := NewWriter()
	arr := w.Array()
	arr.String(`string says "hello"`)
	arr.String("ok")
	arr.End()
	fmt.Println(string(w.Bytes()))
	// Output: ["string says \"hello\"","ok"]
}

func ExampleArrayState_Array() {
	w := NewWriter()
	arr := w.Array()
	arr.Int(1)
	subArr := arr.Array()
	subArr.Int(2)
	subArr.Int(3)
	subArr.End()
	arr.Int(4)
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [1,[2,3],4]
}

func ExampleArrayState_Object() {
	w := NewWriter()
	arr := w.Array()
	obj1 := arr.Object()
	obj1.Int("value", 1)
	obj1.End()
	obj2 := arr.Object()
	obj2.Int("value", 2)
	obj2.End()
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [{"value":1},{"value":2}]
}

func ExampleArrayState_Raw() {
	data := json.RawMessage(`{"value":1}`)
	w := NewWriter()
	arr := w.Array()
	arr.Raw(data)
	arr.End()

	fmt.Println(string(w.Bytes()))
	// Output: [{"value":1}]
}
