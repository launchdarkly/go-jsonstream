package jwriter

import (
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/internal/commontest"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSONWithWriter(t *testing.T) {
	data, err := MarshalJSONWithWriter(ExampleStructWrapper(commontest.ExampleStructValue))
	assert.NoError(t, err)
	assert.Equal(t, commontest.ExampleStructData, data)
}
