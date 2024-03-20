package pkg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// ToCtyValue is a function that converts a primary/collection type to cty.Value
func ToCtyValue(input interface{}) cty.Value {

	val := reflect.ValueOf(input)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cty.NumberIntVal(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cty.NumberUIntVal(val.Uint())
	case reflect.Float32, reflect.Float64:
		return cty.NumberFloatVal(val.Float())
	case reflect.String:
		return cty.StringVal(val.String())
	case reflect.Bool:
		return cty.BoolVal(val.Bool())
	case reflect.Slice:
		if val.Len() == 0 {
			sliceType := reflect.TypeOf(input)
			return cty.ListValEmpty(GoTypeToCtyType(sliceType.Elem()))
		}
		var vals []cty.Value
		for i := 0; i < val.Len(); i++ {
			vals = append(vals, ToCtyValue(val.Index(i).Interface()))
		}
		return cty.ListVal(vals)
	case reflect.Map:
		if val.Len() == 0 {
			mapType := reflect.TypeOf(input)
			elementType := mapType.Elem()
			return cty.MapValEmpty(GoTypeToCtyType(elementType))
		}
		vals := make(map[string]cty.Value)
		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			vals[key] = ToCtyValue(iter.Value().Interface())
		}
		return cty.MapVal(vals)
	case reflect.Struct:
		vals := make(map[string]cty.Value)
		for i := 0; i < val.NumField(); i++ {
			fieldName := val.Type().Field(i).Name
			fieldValue := val.Field(i)
			vals[fieldName] = ToCtyValue(fieldValue.Interface())
		}
		return cty.ObjectVal(vals)
	case reflect.Ptr:
		if val.IsNil() {
			return cty.NilVal
		}
		return ToCtyValue(val.Elem().Interface())
	default:
		return cty.NilVal
	}
}

func GoTypeToCtyType(goType reflect.Type) cty.Type {
	if goType == nil {
		return cty.NilType
	}
	switch goType.Kind() {
	case reflect.Bool:
		return cty.Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cty.Number
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cty.Number
	case reflect.Float32, reflect.Float64:
		return cty.Number
	case reflect.String:
		return cty.String
	case reflect.Slice:
		elemType := GoTypeToCtyType(goType.Elem())
		return cty.List(elemType)
	case reflect.Map:
		valueType := GoTypeToCtyType(goType.Elem())
		return cty.Map(valueType)
	default:
		return cty.NilType
	}
}

func Int(i int) *int {
	return &i
}

func CtyValueToString(val cty.Value) string {
	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		bf := val.AsBigFloat()
		return bf.Text('f', -1)
	case cty.Bool:
		return fmt.Sprintf("%t", val.True())
	case cty.NilType:
		return "nil"
	default:
		if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
			strs := make([]string, 0, val.LengthInt())
			it := val.ElementIterator()
			for it.Next() {
				_, v := it.Element()
				strs = append(strs, CtyValueToString(v))
			}
			return "[" + strings.Join(strs, ", ") + "]"
		} else if val.Type().IsMapType() || val.Type().IsObjectType() {
			strs := make([]string, 0, val.LengthInt())
			it := val.ElementIterator()
			for it.Next() {
				k, v := it.Element()
				strs = append(strs, fmt.Sprintf("%s: %s", k.AsString(), CtyValueToString(v)))
			}
			return "{" + strings.Join(strs, ", ") + "}"
		} else {
			// For other types, use the GoString method, which will give a
			// string representation of the internal structure of the value.
			return val.GoString()
		}
	}
}
