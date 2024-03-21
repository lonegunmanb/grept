package golden

import (
	"context"
	"fmt"
	"github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
)

type IDag interface {
	Dag() *Dag
}

type Config interface {
	IDag
	Context() context.Context
	EvalContext() *hcl.EvalContext
	RunDag(onReady func(Config, *Dag, *linkedlistqueue.Queue, Block) error) error
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
	err = config.Dag().buildDag(blocks)
	if err != nil {
		return err
	}
	err = config.Dag().runDag(config, tryEvalLocal)
	if err != nil {
		return err
	}
	err = config.Dag().runDag(config, expandBlocks)
	if err != nil {
		return err
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
