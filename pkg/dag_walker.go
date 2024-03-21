package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var _ hclsyntax.Walker = dagWalker{}

type dagWalker struct {
	dag          *Dag
	startAddress string
}

func newDagWalker(d *Dag, startAddress string) dagWalker {
	return dagWalker{
		dag:          d,
		startAddress: startAddress,
	}
}

func (d dagWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	diag := hcl.Diagnostics{}
	if expr, ok := node.(hclsyntax.Expression); ok {
		traversals := expr.Variables()
		for _, traversal := range traversals {
			for i, traverser := range traversal {
				name := name(traverser)
				refIter, ok := refIters[name]
				if !ok {
					continue
				}
				for _, src := range refIter(traversal, i) {
					dest := d.startAddress
					dests, err := d.dag.GetChildren(src)
					if err != nil {
						continue
					}

					if _, edgeExist := dests[dest]; !edgeExist {
						err := d.dag.addEdge(src, dest)
						if err != nil {
							diag = diag.Append(&hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "cannot add edge",
								Detail:   err.Error(),
							})
						}
					}
				}
			}
		}
	}
	return diag
}

func (d dagWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	return nil
}
