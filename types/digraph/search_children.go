package digraph

import ()

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/ext"
)

func searchChildren(n *SearchNode) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if nodes, err := precheckChildren(n, dt.ChildCount, dt.Children); err != nil {
		return nil, err
	} else if nodes != nil {
		return nodes, nil
	}
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := ext.NewCollector(dt.MaxVertices)
	add := validExtChecker(dt, func(sg *goiso.SubGraph, e *goiso.Edge) {
		dt.Extender.Extend(sg, e, exts.Ch())
	})
	embeddings, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	added := 0
	for _, sg := range embeddings {
		for i := range sg.V {
			u := &sg.V[i]
			for _, e := range dt.G.Kids[u.Id] {
				added += add(sg, e)
			}
			for _, e := range dt.G.Parents[u.Id] {
				added += add(sg, e)
			}
		}
	}
	return nodesFromEmbeddings(n, exts.Wait(added))
}
