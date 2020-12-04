package jreader

import "fmt"

func ExampleNewReader() {
	r := NewReader([]byte(`"a \"good\" string"`))
	s := r.String()
	if err := r.Error(); err != nil {
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println(s)
	}
	// Output: a "good" string
}

func ExampleReader_RequireEOF() {
	r := NewReader([]byte(`100,"extra"`))
	n := r.Int()
	err := r.RequireEOF()
	fmt.Println(n, err)
	// Output: 100 unexpected data after end of JSON value at position 3
}

func ExampleReader_AddError() {
	r := NewReader([]byte(`[1,2,3,4,5]`))
	values := []int{}
	for arr := r.Array(); arr.Next(); {
		n := r.Int()
		values = append(values, n)
		if n > 1 {
			r.AddError(fmt.Errorf("got an error after %d", n))
		}
	}
	err := r.Error()
	fmt.Println(values, err)
	// Output: [1 2] got an error after 2
}

func ExampleReader_Null() {
	r := NewReader([]byte(`null`))
	if err := r.Null(); err != nil {
		fmt.Println("error:", err)

	} else {
		fmt.Println("got a null")
	}
	// Output: got a null
}

func ExampleReader_Bool() {
	r := NewReader([]byte(`true`))
	var value bool = r.Bool()
	if err := r.Error(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("value:", value)
	}
	// Output: value: true
}

func ExampleReader_BoolOrNull() {
	r1 := NewReader([]byte(`null`))
	if value1, nonNull := r1.BoolOrNull(); nonNull {
		fmt.Println("value1:", value1)
	}
	r2 := NewReader([]byte(`false`))
	if value2, nonNull := r2.BoolOrNull(); nonNull {
		fmt.Println("value2:", value2)
	}
	// Output: value2: false
}

func ExampleReader_Int() {
	r := NewReader([]byte(`123`))
	var value int = r.Int()
	if err := r.Error(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("value:", value)
	}
	// Output: value: 123
}

func ExampleReader_IntOrNull() {
	r1 := NewReader([]byte(`null`))
	if value1, nonNull := r1.IntOrNull(); nonNull {
		fmt.Println("value1:", value1)
	}
	r2 := NewReader([]byte(`0`))
	if value2, nonNull := r2.IntOrNull(); nonNull {
		fmt.Println("value2:", value2)
	}
	// Output: value2: 0
}

func ExampleReader_Float64() {
	r := NewReader([]byte(`1234.5`))
	var value float64 = r.Float64()
	if err := r.Error(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("value:", value)
	}
	// Output: value: 1234.5
}

func ExampleReader_Float64OrNull() {
	r1 := NewReader([]byte(`null`))
	if value1, nonNull := r1.Float64OrNull(); nonNull {
		fmt.Println("value1:", value1)
	}
	r2 := NewReader([]byte(`0`))
	if value2, nonNull := r2.Float64OrNull(); nonNull {
		fmt.Println("value2:", value2)
	}
	// Output: value2: 0
}

func ExampleReader_String() {
	r := NewReader([]byte(`"a \"good\" string"`))
	var value string = r.String()
	if err := r.Error(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("value:", value)
	}
	// Output: value: a "good" string
}

func ExampleReader_StringOrNull() {
	r1 := NewReader([]byte(`null`))
	if value1, nonNull := r1.StringOrNull(); nonNull {
		fmt.Println("value1:", "\""+value1+"\"")
	}
	r2 := NewReader([]byte(`""`))
	if value2, nonNull := r2.StringOrNull(); nonNull {
		fmt.Println("value2:", "\""+value2+"\"")
	}
	// Output: value2: ""
}

func ExampleReader_Array() {
	r := NewReader([]byte(`[1,2]`))
	values := []int{}
	for arr := r.Array(); arr.Next(); {
		values = append(values, r.Int())
	}
	fmt.Println("values:", values)
	// Output: values: [1 2]
}

func ExampleReader_ArrayOrNull() {
	printArray := func(input string) {
		r := NewReader([]byte(input))
		values := []int{}
		arr := r.Array()
		for arr.Next() {
			values = append(values, r.Int())
		}
		fmt.Println(input, "->", values, "... IsDefined =", arr.IsDefined())
	}
	printArray("null")
	printArray("[1,2]")
	// Output: null -> [] ... IsDefined = false
	// [1,2] -> [1 2] ... IsDefined = true
}

func ExampleReader_Object() {
	r := NewReader([]byte(`{"a":1,"b":2}`))
	items := []string{}
	for obj := r.Object(); obj.Next(); {
		name := obj.Name()
		value := r.Int()
		items = append(items, fmt.Sprintf("%s=%d", name, value))
	}
	fmt.Println("items:", items)
	// Output: items: [a=1 b=2]
}

func ExampleReader_ObjectOrNull() {
	printObject := func(input string) {
		r := NewReader([]byte(input))
		items := []string{}
		obj := r.Object()
		for obj.Next() {
			name := obj.Name()
			value := r.Int()
			items = append(items, fmt.Sprintf("%s=%d", name, value))
		}
		fmt.Println(input, "->", items, "... IsDefined =", obj.IsDefined())
	}
	printObject("null")
	printObject(`{"a":1,"b":2}`)
	// Output: null -> [] ... IsDefined = false
	// {"a":1,"b":2} -> [a=1 b=2] ... IsDefined = true
}

func ExampleReader_Any() {
	printValue := func(input string) {
		r := NewReader([]byte(input))
		value := r.Any()
		switch value.Kind {
		case NullValue:
			fmt.Println("a null")
		case BoolValue:
			fmt.Println("a bool:", value.Bool)
		case NumberValue:
			fmt.Println("a number:", value.Number)
		case StringValue:
			fmt.Println("a string:", value.String)
		case ArrayValue:
			n := 0
			for value.Array.Next() {
				n++ // for this example, we're not looking at the actual element value
			}
			fmt.Println("an array with", n, "elements")
		case ObjectValue:
			n := 0
			for value.Object.Next() {
				n++ // for this example, we're not looking at the actual element value
			}
		}
	}
	printValue(`123`)
	printValue(`["a","b"]`)
	// Output: a number: 123
	// an array with 2 elements
}

func ExampleObjectState_WithRequiredProperties() {
	requiredProps := []string{"key", "name"}
	r := NewReader([]byte(`{"name": "x"}`))
	var key, name string
	for obj := r.Object().WithRequiredProperties(requiredProps); obj.Next(); {
		switch string(obj.Name()) {
		case "key":
			key = r.String()
		case "name":
			name = r.String()
		}
	}
	if err := r.Error(); err != nil {
		if rpe, ok := err.(RequiredPropertyError); ok {
			fmt.Println("missing property:", rpe.Name)
		} else {
			fmt.Println("unexpected error:", err)
		}
	} else {
		fmt.Println(key, name)
	}
	// Output: missing property: key
}
