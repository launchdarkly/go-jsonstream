package jwriter

import "fmt"

func Example() {
	w := NewWriter()

	obj := w.Object()
	obj.String("propertyName", "propertyValue")
	obj.End()

	if err := w.Error(); err != nil {
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println(string(w.Bytes()))
	}

	// Output: {"propertyName":"propertyValue"}
}
