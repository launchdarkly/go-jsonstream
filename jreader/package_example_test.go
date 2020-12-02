package jreader

import "fmt"

func Example() {
	r := NewReader([]byte(`"a \"good\" string"`))

	s := r.String()

	if err := r.Error(); err != nil {
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println(s)
	}

	// Output: a "good" string
}
