package golden

import "github.com/hashicorp/hcl/v2/hclsyntax"

type PreCondition struct {
	Body         *hclsyntax.Body
	Condition    bool   `hcl:"condition"`
	ErrorMessage string `hcl:"error_message,optional"`
}
