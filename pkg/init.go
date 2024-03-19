package pkg

func init() {
	registerLocal()
	registerValidator()
	registerRule()
	registerFix()
	registerData()
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
