package pkg

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/lonegunmanb/hclfuncs"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

type IDag interface {
	Dag() *Dag
}

type Config interface {
	IDag
	Context() context.Context
	EvalContext() *hcl.EvalContext
	setDag(*Dag)
}

var _ Config = &GreptConfig{}

type BaseConfig struct {
	ctx     context.Context
	basedir string
	dag     *Dag
}

func (c *BaseConfig) Context() context.Context {
	return c.ctx
}

func (c *BaseConfig) setDag(d *Dag) {
	c.dag = d
}

func Blocks[T Block](c IDag) []T {
	var r []T
	for _, b := range c.Dag().GetVertices() {
		t, ok := b.(T)
		if ok {
			r = append(r, t)
		}
	}
	return r
}

func (c *BaseConfig) Dag() *Dag {
	return c.dag
}

func (c *BaseConfig) blocksByTypes() map[string][]Block {
	r := make(map[string][]Block)
	for _, b := range blocks(c) {
		bt := b.BlockType()
		r[bt] = append(r[bt], b)
	}
	return r
}

func (c *BaseConfig) EvalContext() *hcl.EvalContext {
	ctx := hcl.EvalContext{
		Functions: hclfuncs.Functions(c.basedir),
		Variables: make(map[string]cty.Value),
	}
	for bt, bs := range c.blocksByTypes() {
		sample := bs[0]
		if _, ok := sample.(SingleValueBlock); ok {
			ctx.Variables[bt] = SingleValues(castBlock[SingleValueBlock](bs))
			continue
		}
		ctx.Variables[bt] = Values(bs)
	}
	return &ctx
}

func NewBasicConfig(basedir string, ctx context.Context) *BaseConfig {
	if ctx == nil {
		ctx = context.Background()
	}
	c := &BaseConfig{
		basedir: basedir,
		ctx:     ctx,
	}
	return c
}

func InitConfig(config Config, hclBlocks []*HclBlock) error {
	var err error

	var blocks []Block
	for _, hb := range hclBlocks {
		b, wrapError := wrapBlock(config, hb)
		if wrapError != nil {
			err = multierror.Append(wrapError)
			continue
		}
		blocks = append(blocks, b)
	}
	if err != nil {
		return err
	}
	// If there's dag error, return dag error first.
	dag, err := newDag(blocks)
	if err != nil {
		return err
	}
	config.setDag(dag)
	err = dag.runDag(config, tryEvalLocal)
	if err != nil {
		return err
	}
	err = dag.runDag(config, expandBlocks)
	if err != nil {
		return err
	}

	return nil
}

func RunGreptPlan(c Config) (*GreptPlan, error) {
	err := c.Dag().runDag(c, plan)
	if err != nil {
		return nil, err
	}

	plan := newPlan()
	for _, rb := range Blocks[Rule](c) {
		checkErr := rb.CheckError()
		if checkErr == nil {
			continue
		}
		plan.addRule(&FailedRule{
			Rule:       rb,
			CheckError: checkErr,
		})
		for _, fb := range Blocks[Fix](c) {
			if linq.From(fb.GetRuleIds()).Contains(rb.Id()) {
				plan.addFix(fb)
			}
		}
	}

	return plan, nil
}

func planBlock(b Block) error {
	decodeErr := decode(b)
	if decodeErr != nil {
		return fmt.Errorf("%s.%s.%s(%s) decode error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), decodeErr)
	}
	if validateErr := Validate.Struct(b); validateErr != nil {
		return fmt.Errorf("%s.%s.%s is not valid: %s", b.BlockType(), b.Type(), b.Name(), validateErr.Error())
	}
	failedChecks, preConditionCheckError := b.PreConditionCheck(b.EvalContext())
	if preConditionCheckError != nil {
		return preConditionCheckError
	}
	if len(failedChecks) > 0 {
		var err error
		for _, c := range failedChecks {
			err = multierror.Append(err, fmt.Errorf("precondition check error: %s, %s", c.ErrorMessage, c.Body.Range().String()))
		}
		return err
	}
	pa, ok := b.(PlanBlock)
	if ok {
		execErr := pa.ExecuteDuringPlan()
		if execErr != nil {
			return fmt.Errorf("%s.%s.%s(%s) exec error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), execErr)
		}
	}
	return nil
}

func wrapBlock(c Config, hb *HclBlock) (Block, error) {
	blockFactories := factories[hb.Type]
	blockType := ""
	if len(hb.Labels) > 0 {
		blockType = hb.Labels[0]
	}
	f, ok := blockFactories[blockType]
	if !ok {
		return nil, fmt.Errorf("unregistered %s: %s", hb.Type, blockType)
	}
	return f(c, hb), nil
}

func loadGreptHclBlocks(ignoreUnsupportedBlock bool, dir string) ([]*HclBlock, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
	}

	var blocks []*HclBlock

	for _, filename := range matches {
		content, fsErr := afero.ReadFile(fs, filename)
		if fsErr != nil {
			err = multierror.Append(err, fsErr)
			continue
		}
		file, diag := hclsyntax.ParseConfig(content, filename, hcl.InitialPos)
		if diag.HasErrors() {
			err = multierror.Append(err, diag.Errs()...)
			continue
		}
		body := file.Body.(*hclsyntax.Body)
		for _, b := range body.Blocks {
			var bs []*hclsyntax.Block = readRawHclBlock(b)
			for _, hb := range bs {
				blocks = append(blocks, newHclBlock(hb, nil))
			}
		}
	}
	if err != nil {
		return nil, err
	}

	var r []*HclBlock

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if validBlockTypes.Contains(b.Type) {
			r = append(r, b)
			continue
		}
		if !ignoreUnsupportedBlock {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
		}
	}
	return r, err
}

func readRawHclBlock(b *hclsyntax.Block) []*hclsyntax.Block {
	if b.Type != "locals" {
		return []*hclsyntax.Block{b}
	}
	var newBlocks []*hclsyntax.Block
	for _, attr := range b.Body.Attributes {
		newBlocks = append(newBlocks, &hclsyntax.Block{
			Type:   "local",
			Labels: []string{"", attr.Name},
			Body: &hclsyntax.Body{
				Attributes: map[string]*hclsyntax.Attribute{
					"value": {
						Name:        "value",
						Expr:        attr.Expr,
						SrcRange:    attr.SrcRange,
						NameRange:   attr.NameRange,
						EqualsRange: attr.EqualsRange,
					},
				},
				SrcRange: attr.NameRange,
				EndRange: attr.SrcRange,
			},
		})
	}
	return newBlocks
}

func blocks(c IDag) []Block {
	var blocks []Block
	for _, n := range c.Dag().GetVertices() {
		blocks = append(blocks, n.(Block))
	}
	return blocks
}

func castBlock[T Block](s []Block) []T {
	var r []T
	for _, b := range s {
		r = append(r, b.(T))
	}
	return r
}
