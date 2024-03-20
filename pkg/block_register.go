package pkg

import (
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/lonegunmanb/go-defaults"
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type blockConstructor = func(Config, *hclBlock) Block
type blockRegistry map[string]blockConstructor

var validBlockTypes sets.Set = hashset.New()
var refIters = map[string]refIterator{}

var baseFactory = map[string]func() any{}

func RegisterBaseBlock(factory func() BlockType) {
	bb := factory()
	baseFactory[bb.BlockType()] = func() any {
		return factory()
	}
}

func RegisterBlock(t Block) {
	bt := t.BlockType()
	registry, ok := factories[bt]
	if !ok {
		registry = make(blockRegistry)
		factories[bt] = registry
	}
	_, ok = refIters[bt]
	if !ok {
		refIters[bt] = iterator(bt, t.AddressLength())
	}
	validBlockTypes.Add(bt)
	registry[t.Type()] = func(c Config, hb *hclBlock) Block {
		newBlock := reflect.New(reflect.TypeOf(t).Elem()).Elem()
		newBaseBlock := newBaseBlock(c, hb)
		newBaseBlock.setForEach(hb.forEach)
		newBaseBlock.setMetaNestedBlock()
		newBlock.FieldByName("BaseBlock").Set(reflect.ValueOf(newBaseBlock))
		b := newBlock.Addr().Interface().(Block)
		if f, ok := baseFactory[bt]; ok {
			blockName := cases.Title(language.English).String(bt)
			newBlock.FieldByName("Base" + blockName).Set(reflect.ValueOf(f()))
		}
		defaults.SetDefaults(b)
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
