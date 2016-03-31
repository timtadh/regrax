package digraph

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
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
		return []lattice.Node{NewEmbListNode(dt, nil)}, nil
	}
	if nodes, has, err := cached(n, dt, dt.ParentCount, dt.Parents); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	embs, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	nodes = make([]lattice.Node, 0, 10)
	for _, psg := range embs[0].SubGraphs() {
		psn := NewSubgraphPattern(dt, psg)
		psgs, err := psn.Embeddings()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n.New(psgs))
	}
	if len(nodes) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return nodes, cache(dt, dt.ParentCount, dt.Parents, n.Label(), nodes)
}
