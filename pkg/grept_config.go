package pkg

import (
	"context"
	"fmt"
	"github.com/Azure/grept/golden"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"path/filepath"
)

var _ golden.Config = &GreptConfig{}

type GreptConfig struct {
	*golden.BaseConfig
}

func NewGreptConfig(baseDir string, ctx context.Context, hclBlocks []*golden.HclBlock) (golden.Config, error) {
	cfg := &GreptConfig{
		BaseConfig: golden.NewBasicConfig(baseDir, ctx),
	}
	return cfg, golden.InitConfig(cfg, hclBlocks)
}

func BuildGreptConfig(baseDir, cfgDir string, ctx context.Context) (golden.Config, error) {
	var err error
	hclBlocks, err := loadGreptHclBlocks(false, cfgDir)
	if err != nil {
		return nil, err
	}

	c, err := NewGreptConfig(baseDir, ctx, hclBlocks)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loadGreptHclBlocks(ignoreUnsupportedBlock bool, dir string) ([]*golden.HclBlock, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
	}

	var blocks []*golden.HclBlock

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
				blocks = append(blocks, golden.NewHclBlock(hb, nil))
			}
		}
	}
	if err != nil {
		return nil, err
	}

	var r []*golden.HclBlock

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if golden.IsBlockTypeWanted(b.Type) {
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
