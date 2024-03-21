package pkg

import (
	"github.com/emirpasic/gods/queues/linkedlistqueue"
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

func newDag() *Dag {
	return &Dag{
		DAG:              dag.NewDAG(),
		pendingUpstreams: make(map[string]sets.Set),
	}
}

func (d *Dag) buildDag(blocks []Block) error {
	//g := &Dag{
	//	DAG:              dag.NewDAG(),
	//	pendingUpstreams: make(map[string]sets.Set),
	//}
	var walkErr error
	for _, b := range blocks {
		err := d.AddVertexByID(blockAddress(b.HclBlock()), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.HclBlock().Body, dagWalker{dag: d, rootBlock: b})
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return walkErr
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

func (d *Dag) runDag(c Config, onReady func(Config, *Dag, *linkedlistqueue.Queue, Block) error) error {
	var err error
	visited := hashset.New()
	pending := linkedlistqueue.New()
	for _, n := range d.GetRoots() {
		pending.Enqueue(n.(Block))
	}
	for !pending.Empty() {
		next, _ := pending.Dequeue()
		b := next.(Block)
		// the node has already been expanded and deleted from dag
		address := blockAddress(b.HclBlock())
		exist := d.exist(address)
		if !exist {
			continue
		}
		ancestors, dagErr := d.GetAncestors(address)
		if dagErr != nil {
			return dagErr
		}
		ready := true
		for upstreamAddress := range ancestors {
			if !visited.Contains(upstreamAddress) {
				ready = false
			}
		}
		if !ready {
			continue
		}
		if callbackErr := onReady(c, d, pending, b); callbackErr != nil {
			err = multierror.Append(err, callbackErr)
		}
		visited.Add(address)
		// this address might be expanded during onReady and no more exist.
		exist = d.exist(address)
		if !exist {
			continue
		}
		children, dagErr := d.GetChildren(address)
		if dagErr != nil {
			return dagErr
		}
		for _, n := range children {
			pending.Enqueue(n)
		}
	}
	return err
}

func (d *Dag) exist(address string) bool {
	n, existErr := d.GetVertex(address)
	notExist := n == nil || existErr != nil
	return !notExist
}
