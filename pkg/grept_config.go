package pkg

type GreptConfig struct {
	*BaseConfig
}

func (g *GreptConfig) IgnoreUnsupportedBlock() bool {
	return false
}

func NewGreptConfig() *GreptConfig {
	return &GreptConfig{
		BaseConfig: NewBasicConfig(),
	}
}
