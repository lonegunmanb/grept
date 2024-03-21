package golden

import "github.com/zclconf/go-cty/cty"

type Valuable interface {
	Values() map[string]cty.Value
}
