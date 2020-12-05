package commontest

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ExampleStruct struct {
	StringField             string      `json:"string"`
	IntField                int         `json:"int"`
	OptBoolAsInterfaceField interface{} `json:"optBool"`
}

const (
	ExampleStructStringFieldName             = "string"
	ExampleStructIntFieldName                = "int"
	ExampleStructOptBoolAsInterfaceFieldName = "optBool"
)

var (
	ExampleStructData               = []byte(`{"string":"s","int":3,"optBool":true}`)
	ExampleStructValue              = ExampleStruct{StringField: "s", IntField: 3, OptBoolAsInterfaceField: true}
	ExampleStructRequiredFieldNames = []string{ExampleStructStringFieldName, ExampleStructIntFieldName}
)

func MakeBools() []bool {
	ret := make([]bool, 0, 100)
	for i := 0; i < 50; i++ {
		ret = append(ret, false, true)
	}
	return ret
}

func MakeBoolsJSON(bools []bool) []byte {
	var buf bytes.Buffer
	buf.WriteRune('[')
	for i, val := range bools {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(fmt.Sprintf("%t", val))
	}
	buf.WriteRune(']')
	return buf.Bytes()
}

func MakeStrings() []string {
	ret := make([]string, 0, 100)
	for i := 0; i < 50; i++ {
		ret = append(ret, fmt.Sprintf("value%d", i))
		ret = append(ret, fmt.Sprintf("value\twith\n\"escaped chars\"%d", i))
	}
	return ret
}

func MakeStringsJSON(strings []string) []byte {
	var buf bytes.Buffer
	buf.WriteRune('[')
	for i, val := range strings {
		if i > 0 {
			buf.WriteRune(',')
		}
		data, _ := json.Marshal(val)
		_, _ = buf.Write(data)
	}
	buf.WriteRune(']')
	return buf.Bytes()
}

func MakeStructs() []ExampleStruct {
	ret := make([]ExampleStruct, 0, 100)
	for i := 0; i < 100; i++ {
		ret = append(ret, ExampleStruct{
			StringField:             fmt.Sprintf("string%d", i),
			IntField:                i * 10,
			OptBoolAsInterfaceField: i%2 == 1,
		})
	}
	return ret
}

func MakeStructsJSON(structs []ExampleStruct) []byte {
	bytes, _ := json.Marshal(structs)
	return bytes
}
