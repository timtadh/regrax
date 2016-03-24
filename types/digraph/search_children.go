package digraph

import ()

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/ext"
	"github.com/timtadh/sfp/types/digraph/support"
)

func searchChildren(n *SearchNode) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if n.isRoot() {
		return n.loadFrequentVertices()
	}
	if n.edges() >= dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := cached(n, dt, dt.ChildCount, dt.Children); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := ext.NewCollector(dt.MaxVertices)
	add := func(sg *goiso.SubGraph, e *goiso.Edge) int {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return 0
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			dt.Extender.Extend(sg, e, exts.Ch())
			return 1
		}
		return 0
	}
	embeddings, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	added := 0
	sup, err := dt.Supported(dt, embeddings)
	if err != nil {
		return nil, err
	}
	sizes := set.NewSortedSet(len(embeddings[0].V))
	for _, set := range support.VertexMapSets(embeddings) {
		sizes.Add(types.Int(set.Size()))
	}
	errors.Logf("SEARCH-EMBEDDINGS", "len(V) %v len(embeddings) %v supported %v unique-vertex-embeddings %v", len(embeddings[0].V), len(embeddings), len(sup), sizes)
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
	exts.Wait(added)
	partitioned := exts.Partition()
	sum := 0
	for _, sgs := range partitioned {
		sum += len(sgs)
		new_node := n.New(support.Dedup(sgs))
		if len(sgs) < dt.Support() {
			continue
		}
		new_embeddings, err := new_node.Embeddings()
		if err != nil {
			return nil, err
		}
		supported, err := dt.Supported(dt, new_embeddings)
		if err != nil {
			return nil, err
		}
		if len(supported) >= dt.Support() {
			nodes = append(nodes, new_node)
		}
	}
	// errors.Logf("DEBUG", "sum(len(partition)) %v", sum)
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, cache(dt, dt.ChildCount, dt.Children, n.Label(), nodes)
}
