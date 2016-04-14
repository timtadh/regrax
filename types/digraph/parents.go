package digraph

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/subgraph"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

func parents(n Node, parents bytes_bytes.MultiMap, parentCount bytes_int.MultiMap) (nodes []lattice.Node, err error) {
	// errors.Logf("DEBUG", "compute Parents\n    of %v", n)
	if n.isRoot() {
		return []lattice.Node{}, nil
	}
	dt := n.dt()
	sg := n.SubGraph()
	if len(sg.V) == 1 && len(sg.E) == 0 {
		return []lattice.Node{NewEmbListNode(dt, nil, nil)}, nil
	}
	if nodes, has, err := cached(n, dt, dt.ParentCount, dt.Parents); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	emb, err := n.Embedding()
	if err != nil {
		return nil, err
	}
	nodes = make([]lattice.Node, 0, 10)
	for _, parent := range emb.SubGraphs() {
		psg := subgraph.FromEmbedding(parent)
		pexts, pembs, err := extsAndEmbs(dt, psg)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n.New(pexts, pembs))
	}
	if len(nodes) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return nodes, cache(dt, dt.ParentCount, dt.Parents, n.Label(), nodes)
}
