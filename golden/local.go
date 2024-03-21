package golden

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Local = &LocalBlock{}

type Local interface {
	SingleValueBlock
	// discriminator func
	Local()
}

type LocalBlock struct {
	*BaseBlock
	LocalValue cty.Value `hcl:"value"`
}

func (l *LocalBlock) Value() cty.Value {
	return l.LocalValue
}

func (l *LocalBlock) Type() string {
	return ""
}

func (l *LocalBlock) BlockType() string {
	return "local"
}

func (l *LocalBlock) Local() {}

func (l *LocalBlock) AddressLength() int { return 2 }
