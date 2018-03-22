package digraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
)

import (
	"github.com/timtadh/regrax/lattice"
)

func parents(n Node) (nodes []lattice.Node, err error) {
	// errors.Logf("DEBUG", "compute Parents\n    of %v", n)
	if n.isRoot() {
		return []lattice.Node{}, nil
	}
	dt := n.dt()
	sg := n.SubGraph()
	if len(sg.V) == 1 && len(sg.E) == 0 {
		return []lattice.Node{dt.Root()}, nil
	}
	parentBuilders, err := AllParents(n.SubGraph().Builder())
	if err != nil {
		return nil, err
	}
	seen := set.NewSortedSet(10)
	nodes = make([]lattice.Node, 0, 10)
	for _, pBuilder := range parentBuilders {
		parent := pBuilder.Build()
		if seen.Has(parent) {
			continue
		}
		seen.Add(parent)
		support, pexts, pembs, poverlap, err := ExtsAndEmbs(dt, parent, nil, set.NewSortedSet(0), dt.Mode, false)
		if err != nil {
			return nil, err
		}
		if support < dt.Support() {
			// this means this parent support comes from automorphism
			// it isn't truly supported, and its children may be spurious as well
			// log and skip?
			ExtsAndEmbs(dt, parent, nil, set.NewSortedSet(0), dt.Mode, true)

			errors.Logf("WARN", "for node %v parent %v had support %v less than required %v due to automorphism", n, parent.Pretty(dt.Labels), support, dt.Support())
		} else {
			nodes = append(nodes, n.New(parent, pexts, pembs, poverlap))
		}
	}
	if len(nodes) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return nodes, nil
}
