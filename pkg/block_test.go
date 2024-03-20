package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
	"math/big"
	"testing"
)

func Test_LocalBlocksValueShouldBeAFlattenObject(t *testing.T) {
	numberVal := cty.NumberVal(big.NewFloat(1))
	stringVal := cty.StringVal("hello world")
	locals := []Block{
		&LocalBlock{
			BaseBlock: &BaseBlock{
				name: "number_value",
			},
			LocalValue: numberVal,
		},
		&LocalBlock{
			BaseBlock: &BaseBlock{
				name: "string_value",
			},
			LocalValue: stringVal,
		},
	}

	values := SingleValues(castBlock[SingleValueBlock](locals))
	assert.True(t, AreCtyValuesEqual(numberVal, values.GetAttr("number_value")))
	assert.True(t, AreCtyValuesEqual(stringVal, values.GetAttr("string_value")))
}
