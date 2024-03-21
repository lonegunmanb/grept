package pkg

import (
	"github.com/Azure/grept/golden"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Fix interface {
	golden.ApplyBlock
	GetRuleIds() []string
	// discriminator func
	Fix()
	setRuleIds([]string)
}

var _ golden.Valuable = &BaseFix{}
var _ golden.CustomDecodeBase = &BaseFix{}

type BaseFix struct {
	RuleIds []string `json:"rule_ids" hcl:"rule_ids"`
}

func (bf *BaseFix) Fix() {}

func (bf *BaseFix) Decode(hb *golden.HclBlock, evalContext *hcl.EvalContext) error {
	ruleIds, diag := hb.Body.Attributes["rule_ids"].Expr.Value(evalContext)
	if diag.HasErrors() {
		return diag
	}
	var ids []string
	for _, id := range ruleIds.AsValueSlice() {
		ids = append(ids, id.AsString())
	}
	bf.setRuleIds(ids)
	return nil
}

func (bf *BaseFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_ids": golden.ToCtyValue(bf.RuleIds),
	}
}

func (bf *BaseFix) BlockType() string {
	return "fix"
}

func (bf *BaseFix) GetRuleIds() []string {
	return bf.RuleIds
}

func (bf *BaseFix) AddressLength() int { return 3 }

func (bf *BaseFix) setRuleIds(ids []string) {
	bf.RuleIds = ids
}
