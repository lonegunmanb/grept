package pkg

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestToCtyValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  cty.Value
	}{
		{
			name:  "int",
			input: 10,
			want:  cty.NumberIntVal(10),
		},
		{
			name:  "float",
			input: 12.34,
			want:  cty.NumberFloatVal(12.34),
		},
		{
			name:  "string",
			input: "hello",
			want:  cty.StringVal("hello"),
		},
		{
			name:  "slice of number",
			input: []int{1, 2, 3},
			want:  cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)}),
		},
		{
			name:  "slice of string",
			input: []string{"1", "2", "3"},
			want:  cty.ListVal([]cty.Value{cty.StringVal("1"), cty.StringVal("2"), cty.StringVal("3")}),
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  cty.ListValEmpty(cty.String),
		},
		{
			name:  "map of number",
			input: map[string]int{"one": 1, "two": 2},
			want:  cty.MapVal(map[string]cty.Value{"one": cty.NumberIntVal(1), "two": cty.NumberIntVal(2)}),
		},
		{
			name:  "map of bool",
			input: map[string]bool{"one": true, "two": false},
			want:  cty.MapVal(map[string]cty.Value{"one": cty.BoolVal(true), "two": cty.BoolVal(false)}),
		},
		{
			name:  "empty map",
			input: map[string]int{},
			want:  cty.MapValEmpty(cty.Number),
		},
		{
			name:  "nil",
			input: nil,
			want:  cty.NilVal,
		},
		{
			name:  "struct",
			input: TestStruct{Name: "John", Age: 30},
			want: cty.ObjectVal(map[string]cty.Value{
				"Name": cty.StringVal("John"),
				"Age":  cty.NumberIntVal(30),
			}),
		},
		{
			name:  "pointer to number",
			input: Int(1),
			want:  cty.NumberIntVal(1),
		},
		{
			name:  "pointer to struct",
			input: &TestStruct{Name: "John", Age: 30},
			want: cty.ObjectVal(map[string]cty.Value{
				"Name": cty.StringVal("John"),
				"Age":  cty.NumberIntVal(30),
			}),
		},
		{
			name:  "nil pointer to struct",
			input: (*TestStruct)(nil),
			want:  cty.NilVal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCtyValue(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToCtyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoTypeToCtyType(t *testing.T) {
	tests := []struct {
		name        string
		goType      reflect.Type
		wantCtyType cty.Type
	}{
		{
			name:        "int",
			goType:      reflect.TypeOf(int(0)),
			wantCtyType: cty.Number,
		},
		{
			name:        "float",
			goType:      reflect.TypeOf(float64(0)),
			wantCtyType: cty.Number,
		},
		{
			name:        "string",
			goType:      reflect.TypeOf(""),
			wantCtyType: cty.String,
		},
		{
			name:        "bool",
			goType:      reflect.TypeOf(true),
			wantCtyType: cty.Bool,
		},
		{
			name:        "slice of number",
			goType:      reflect.TypeOf([]int{}),
			wantCtyType: cty.List(cty.Number),
		},
		{
			name:        "slice of string",
			goType:      reflect.TypeOf([]string{}),
			wantCtyType: cty.List(cty.String),
		},
		{
			name:        "map of number",
			goType:      reflect.TypeOf(map[string]int{}),
			wantCtyType: cty.Map(cty.Number),
		},
		{
			name:        "map of string",
			goType:      reflect.TypeOf(map[string]string{}),
			wantCtyType: cty.Map(cty.String),
		},
		{
			name:        "nil",
			goType:      reflect.TypeOf(nil),
			wantCtyType: cty.NilType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GoTypeToCtyType(tt.goType); got != tt.wantCtyType {
				t.Errorf("GoTypeToCtyType() = %v, want %v", got, tt.wantCtyType)
			}
		})
	}
}
