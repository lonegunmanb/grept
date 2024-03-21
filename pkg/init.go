package pkg

import "github.com/Azure/grept/golden"

func init() {
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseRule)
	})
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseData)
	})
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseFix)
	})
	registerRule()
	registerFix()
	registerData()
}

func registerFix() {
	golden.RegisterBlock(new(CopyFileFix))
	golden.RegisterBlock(new(LocalFileFix))
	golden.RegisterBlock(new(RenameFileFix))
	golden.RegisterBlock(new(RmLocalFileFix))
	golden.RegisterBlock(new(LocalShellFix))
	golden.RegisterBlock(new(GitIgnoreFix))
	golden.RegisterBlock(new(YamlTransformFix))
}

func registerRule() {
	golden.RegisterBlock(new(FileExistRule))
	golden.RegisterBlock(new(FileHashRule))
	golden.RegisterBlock(new(MustBeTrueRule))
	golden.RegisterBlock(new(DirExistRule))
}

func registerData() {
	golden.RegisterBlock(new(HttpDatasource))
	golden.RegisterBlock(new(GitIgnoreDatasource))
}
