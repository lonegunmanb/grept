package golden

import (
	"reflect"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

func Value(b Block) map[string]cty.Value {
	values := make(map[string]cty.Value)

	blockType := reflect.TypeOf(b)
	blockValue := reflect.ValueOf(b)

	// Check if the block is a pointer and dereference it if so
	if blockType.Kind() == reflect.Ptr {
		blockType = blockType.Elem()
		blockValue = blockValue.Elem()
	}

	for i := 0; i < blockType.NumField(); i++ {
		field := blockType.Field(i)
		tagName, tagDefined := fieldName(field)
		if !tagDefined {
			continue
		}
		fieldValue := blockValue.Field(i)
		values[tagName] = ToCtyValue(fieldValue.Interface())
	}

	return values
}

func fieldName(f reflect.StructField) (string, bool) {
	attributeTag := f.Tag.Get("attribute")

	tag := f.Tag.Get("hcl")
	if attributeTag != "" {
		tag = attributeTag
	}
	if tag == "" {
		return f.Name, false
	}
	tagSegments := strings.Split(tag, ",")
	if len(tagSegments) > 1 {
		tag = tagSegments[0]
	}
	return tag, true
}
