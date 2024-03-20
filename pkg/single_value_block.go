package pkg

import "github.com/zclconf/go-cty/cty"

type SingleValueBlock interface {
	Block
	Value() cty.Value
}
