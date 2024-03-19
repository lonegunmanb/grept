package pkg

import "github.com/hashicorp/hcl/v2"

type GreptConfig struct {
	*BaseConfig
}

func (g *GreptConfig) EvalContext() *hcl.EvalContext {
	ctx := g.BaseConfig.EvalContext()
	ctx.Variables["data"] = Values(Blocks[Data](g))
	ctx.Variables["rule"] = Values(Blocks[Rule](g))
	return ctx
}

func (g *GreptConfig) IgnoreUnsupportedBlock() bool {
	return false
}

func NewGreptConfig() *GreptConfig {
	return &GreptConfig{
		BaseConfig: NewBasicConfig(),
	}
}
