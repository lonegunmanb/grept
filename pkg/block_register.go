package pkg

import (
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type blockConstructor = func(*Config, *hclBlock) block
type blockRegistry map[string]blockConstructor

var baseFactory = map[string]func() any{}

func RegisterBaseBlock(factory func() BlockType) {
	bb := factory()
	baseFactory[bb.BlockType()] = func() any {
		return factory()
	}
}

func RegisterBlock(t block) {
	registry, ok := factories[t.BlockType()]
	if !ok {
		registry = make(blockRegistry)
		factories[t.BlockType()] = registry
	}
	registry[t.Type()] = func(c *Config, hb *hclBlock) block {
		newBlock := reflect.New(reflect.TypeOf(t).Elem()).Elem()
		newBaseBlock := newBaseBlock(c, hb)
		newBaseBlock.setForEach(hb.forEach)
		newBaseBlock.setMetaNestedBlock()
		newBlock.FieldByName("BaseBlock").Set(reflect.ValueOf(newBaseBlock))
		b := newBlock.Addr().Interface().(block)
		if f, ok := baseFactory[t.BlockType()]; ok {
			blockName := cases.Title(language.English).String(t.BlockType())
			newBlock.FieldByName("Base" + blockName).Set(reflect.ValueOf(f()))
		}
		return b
	}
}

func registerLocal() {
	RegisterBlock(new(LocalBlock))
}

var factories = map[string]blockRegistry{}

func registerFix() {
	RegisterBlock(new(CopyFileFix))
	RegisterBlock(new(LocalFileFix))
	RegisterBlock(new(RenameFileFix))
	RegisterBlock(new(RmLocalFileFix))
	RegisterBlock(new(LocalShellFix))
	RegisterBlock(new(GitIgnoreFix))
	RegisterBlock(new(YamlTransformFix))
}

func registerRule() {
	RegisterBlock(new(FileExistRule))
	RegisterBlock(new(FileHashRule))
	RegisterBlock(new(MustBeTrueRule))
	RegisterBlock(new(DirExistRule))
}

func registerData() {
	RegisterBlock(new(HttpDatasource))
	RegisterBlock(new(GitIgnoreDatasource))
}
