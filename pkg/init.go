package pkg

func init() {
	registerRule()
	registerFix()
	registerData()
	registerLocal()
	registerValidator()
	RegisterBaseBlock(func() BlockType {
		return new(BaseRule)
	})
	RegisterBaseBlock(func() BlockType {
		return new(BaseData)
	})
	RegisterBaseBlock(func() BlockType {
		return new(BaseFix)
	})
}
