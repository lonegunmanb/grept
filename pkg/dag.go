package pkg

import (
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
)

type Dag struct {
	*dag.DAG
	pendingUpstreams map[string]sets.Set
}

func newDag(blocks []Block) (*Dag, error) {
	g := &Dag{
		DAG:              dag.NewDAG(),
		pendingUpstreams: make(map[string]sets.Set),
	}
	var walkErr error
	for _, b := range blocks {
		err := g.AddVertexByID(blockAddress(b.HclBlock()), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.HclBlock().Body, dagWalker{dag: g, rootBlock: b})
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return g, walkErr
}

func (d *Dag) addEdge(from, to string) error {
	err := d.AddEdge(from, to)
	if err != nil {
		return err
	}
	set, ok := d.pendingUpstreams[to]
	if !ok {
		set = hashset.New()
		d.pendingUpstreams[to] = set
	}
	set.Add(from)
	return nil
}
