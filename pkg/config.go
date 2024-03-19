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

type Config interface {
	Context() context.Context
	EvalContext() *hcl.EvalContext
	Dag() *Dag
}

var _ Config = &BaseConfig{}

type BaseConfig struct {
	ctx     context.Context
	basedir string
	dag     *Dag
}

func (c *BaseConfig) Context() context.Context {
	return c.ctx
}

type GreptConfig struct {
	*BaseConfig
}

func Blocks[T Block](c *BaseConfig) []T {
	var r []T
	for _, b := range c.dag.GetVertices() {
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

func (c *BaseConfig) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: hclfuncs.Functions(c.basedir),
		Variables: map[string]cty.Value{
			"data":  Values(Blocks[Data](c)),
			"rule":  Values(Blocks[Rule](c)),
			"local": LocalsValues(Blocks[Local](c)),
		},
	}
}

func newEmptyConfig() *BaseConfig {
	c := &BaseConfig{
		ctx: context.TODO(),
	}
	return c
}

func NewConfig(baseDir, cfgDir string, ctx context.Context) (*BaseConfig, error) {
	var err error
	hclBlocks, err := loadHclBlocks(cfgDir)
	if err != nil {
		return nil, err
	}

	c, err := newConfig(baseDir, ctx, hclBlocks, err)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newConfig(baseDir string, ctx context.Context, hclBlocks []*hclBlock, err error) (*BaseConfig, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	config := newEmptyConfig()
	config.basedir = baseDir
	config.ctx = ctx

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
		return nil, err
	}
	// If there's dag error, return dag error first.
	dag, err := newDag(blocks)
	if err != nil {
		return nil, err
	}
	config.dag = dag
	err = dag.runDag(config, tryEvalLocal)
	if err != nil {
		return nil, err
	}
	err = dag.runDag(config, expandBlocks)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *BaseConfig) Plan() (*GreptPlan, error) {
	err := c.dag.runDag(c, plan)
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
	//TODO: Remove this
	if v, ok := b.(Validatable); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("%s.%s.%s is not valid: %s", b.BlockType(), b.Type(), b.Name(), err.Error())
		}
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
	pa, ok := b.(planAction)
	if ok {
		execErr := pa.ExecuteDuringPlan()
		if execErr != nil {
			return fmt.Errorf("%s.%s.%s(%s) exec error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), execErr)
		}
	}
	return nil
}

func wrapBlock(c *BaseConfig, hb *hclBlock) (Block, error) {
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

func loadHclBlocks(dir string) ([]*hclBlock, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
	}

	var blocks []*hclBlock

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

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if !validBlockTypes.Contains(b.Type) {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
			continue
		}
	}
	return blocks, err
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

func (c *BaseConfig) blocks() []Block {
	var blocks []Block
	for _, n := range c.dag.GetVertices() {
		blocks = append(blocks, n.(Block))
	}
	return blocks
}
