package pkg

import "context"

type GreptConfig struct {
	*BaseConfig
}

func NewGreptConfig(baseDir string, ctx context.Context, hclBlocks []*HclBlock) (Config, error) {
	cfg := &GreptConfig{
		BaseConfig: NewBasicConfig(baseDir, ctx),
	}
	return cfg, InitConfig(cfg, hclBlocks)
}

func BuildGreptConfig(baseDir, cfgDir string, ctx context.Context) (Config, error) {
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
