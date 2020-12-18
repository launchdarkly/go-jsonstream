package jwriter

import "fmt"

func Example() {
	w := NewWriter()

	obj := w.Object()
	obj.Name("propertyName").String("propertyValue")
	obj.End()

	if err := w.Error(); err != nil {
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println(string(w.Bytes()))
	}

	// Output: {"propertyName":"propertyValue"}
}
