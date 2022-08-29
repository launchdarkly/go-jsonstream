package jwriter

import (
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/internal/commontest"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSONWithWriter(t *testing.T) {
	data, err := MarshalJSONWithWriter(ExampleStructWrapper(commontest.ExampleStructValue))
	assert.NoError(t, err)
	assert.Equal(t, commontest.ExampleStructData, data)
}
