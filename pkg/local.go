package pkg

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Local = &LocalBlock{}

type Local interface {
	Block
	// discriminator func
	Local()
}

type LocalBlock struct {
	*BaseBlock
	Value cty.Value `hcl:"value"`
}

func (l *LocalBlock) Type() string {
	return ""
}

func (l *LocalBlock) BlockType() string {
	return "local"
}

func (l *LocalBlock) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"": l.Value,
	}
}

func (l *LocalBlock) Local() {}

func (l *LocalBlock) AddressLength() int { return 2 }
