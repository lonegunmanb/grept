package golden

import "github.com/hashicorp/hcl/v2"

type CustomDecodeBase interface {
	Decode(*HclBlock, *hcl.EvalContext) error
}
