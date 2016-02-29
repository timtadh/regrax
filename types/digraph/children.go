package digraph

import (
)

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)


type Node interface {
	lattice.Node
	New([]*goiso.SubGraph) Node
	Label() []byte
	Embeddings() ([]*goiso.SubGraph, error)
	loadFrequentVertices() ([]lattice.Node, error)
	isRoot() bool
	edges() int
	dt() *Digraph
}


func children(n Node, checkCanon bool, children bytes_bytes.MultiMap, childCount bytes_int.MultiMap) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if n.isRoot() {
		return n.loadFrequentVertices()
	}
	if n.edges() >= dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := cached(dt, childCount, children, n.Label()); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) (SubGraphs, error) {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return exts, nil
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return exts, nil
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return exts, nil
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			ext, _ := sg.EdgeExtend(e)
			if len(ext.V) > dt.MaxVertices {
				return exts, nil
			}
			exts = append(exts, ext)
		}
		return exts, nil
	}
	embeddings, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	for _, sg := range embeddings {
		for _, u := range sg.V {
			for _, e := range dt.G.Kids[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
			for _, e := range dt.G.Parents[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	// errors.Logf("DEBUG", "len(exts) %v", len(exts))
	partitioned := exts.Partition()
	sum := 0
	for _, sgs := range partitioned {
		sum += len(sgs)
		new_node := n.New(Dedup(sgs))
		if len(sgs) < dt.Support() {
			continue
		}
		new_embeddings, err := new_node.Embeddings()
		if err != nil {
			return nil, err
		}
		supported := dt.Supported(new_embeddings)
		if len(supported) >= dt.Support() {
			if checkCanon {
				if canonized, err := isCanonicalExtension(embeddings[0], sgs[0]); err != nil {
					return nil, err
				} else if !canonized {
					// errors.Logf("DEBUG", "%v is not canon (skipping)", sgs[0].Label())
				} else {
					nodes = append(nodes, new_node)
				}
			} else {
				nodes = append(nodes, new_node)
			}
		}
	}
	// errors.Logf("DEBUG", "sum(len(partition)) %v", sum)
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, cache(dt, childCount, children, n.Label(), nodes)
}

