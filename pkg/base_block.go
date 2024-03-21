package pkg

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type BaseBlock struct {
	c             Config
	hb            *HclBlock
	name          string
	id            string
	blockAddress  string
	forEach       *forEach
	preConditions []PreCondition
}

func newBaseBlock(c Config, hb *HclBlock) *BaseBlock {
	bb := &BaseBlock{
		c:            c,
		hb:           hb,
		blockAddress: blockAddress(hb),
		name:         hb.Labels[1],
		id:           uuid.NewString(),
	}
	return bb
}

func (bb *BaseBlock) Id() string {
	if bb == nil {
		return ""
	}
	return bb.id
}

func (bb *BaseBlock) Name() string {
	if bb == nil {
		return ""
	}
	return bb.name
}

func (bb *BaseBlock) HclBlock() *HclBlock {
	if bb.hb == nil {
		return &HclBlock{
			Block: new(hclsyntax.Block),
		}
	}
	return bb.hb
}

func (bb *BaseBlock) BaseValues() map[string]cty.Value {
	return map[string]cty.Value{
		"id": cty.StringVal(bb.id),
	}
}

func (bb *BaseBlock) EvalContext() *hcl.EvalContext {
	var ctx *hcl.EvalContext
	if bb.c == nil {
		ctx = new(hcl.EvalContext)
	} else {
		ctx = bb.c.EvalContext()
	}
	if bb.forEach != nil {
		ctx = ctx.NewChild()
		ctx.Variables = map[string]cty.Value{
			"each": cty.ObjectVal(map[string]cty.Value{
				"key":   cty.StringVal(CtyValueToString(bb.forEach.key)),
				"value": bb.forEach.value,
			}),
		}
	}
	return ctx
}

func (bb *BaseBlock) Address() string {
	if bb == nil {
		return ""
	}
	return bb.blockAddress
}

func (bb *BaseBlock) Context() context.Context {
	if bb == nil || bb.c == nil {
		return context.TODO()
	}
	return bb.c.Context()
}

func (bb *BaseBlock) PreConditionCheck(ctx *hcl.EvalContext) ([]PreCondition, error) {
	var failedChecks []PreCondition
	var err error
	for _, cond := range bb.preConditions {
		diag := gohcl.DecodeBody(cond.Body, ctx, &cond)
		if diag.HasErrors() {
			err = multierror.Append(err, diag.Errs()...)
			continue
		}
		if !cond.Condition {
			failedChecks = append(failedChecks, cond)
		}
	}
	return failedChecks, err
}

func (bb *BaseBlock) forEachDefined() bool {
	_, forEach := bb.HclBlock().Body.Attributes["for_each"]
	return forEach
}

func (bb *BaseBlock) getDownstreams() []Block {
	var blocks []Block
	children, _ := bb.c.Dag().GetChildren(bb.blockAddress)
	for _, c := range children {
		blocks = append(blocks, c.(Block))
	}
	return blocks
}

func (bb *BaseBlock) setForEach(each *forEach) {
	bb.forEach = each
}
func (bb *BaseBlock) getForEach() *forEach {
	return bb.forEach
}

func (bb *BaseBlock) setMetaNestedBlock() {
	for _, nb := range bb.hb.Block.Body.Blocks {
		if nb.Type == "precondition" {
			bb.preConditions = append(bb.preConditions, PreCondition{
				Body: nb.Body,
			})
		}
	}
}
