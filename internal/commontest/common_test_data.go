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
	ExampleStructData  = []byte(`{"string":"s","int":3,"optBool":true}`)
	ExampleStructValue = ExampleStruct{StringField: "s", IntField: 3, OptBoolAsInterfaceField: true}
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
