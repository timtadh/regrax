package digraph

import (
	"bytes"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/ext"
	"github.com/timtadh/sfp/types/digraph/subgraph"
	"github.com/timtadh/sfp/types/digraph/support"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

type Node interface {
	lattice.Node
	New([]*goiso.SubGraph) Node
	Label() []byte
	Embeddings() ([]*goiso.SubGraph, error)
	Embedding() (*goiso.SubGraph, error)
	SubGraph() *subgraph.SubGraph
	loadFrequentVertices() ([]lattice.Node, error)
	isRoot() bool
	edges() int
	dt() *Digraph
}

func validExtChecker(dt *Digraph, do func(sg *goiso.SubGraph, e *goiso.Edge)) func(*goiso.SubGraph, *goiso.Edge) int {
	return func(sg *goiso.SubGraph, e *goiso.Edge) int {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return 0
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			do(sg, e)
			return 1
		}
		return 0
	}
}

func precheckChildren(n *EmbListNode, kidCount bytes_int.MultiMap, kids bytes_bytes.MultiMap) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if n.isRoot() {
		return n.loadFrequentVertices()
	}
	if n.edges() >= dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := cached(n, dt, kidCount, kids); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	return nil, nil
}

func nodesFromEmbeddings(n *EmbListNode, embs ext.Embeddings) (nodes []lattice.Node, err error) {
	dt := n.dt()
	partitioned := embs.Partition()
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

func canonChildren(n *EmbListNode) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if nodes, has, err := cached(n, dt, dt.CanonKidCount, dt.CanonKids); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	kids, err := children(n)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(n.Label(), dt.Root().Pattern().Label()) {
		return kids, cache(dt, dt.CanonKidCount, dt.CanonKids, n.Label(), kids)
	}
	nEmb, err := n.Embedding()
	if err != nil {
		return nil, err
	}
	for _, k := range kids {
		kEmb, err := k.(Node).Embedding()
		if err != nil {
			return nil, err
		}
		if canonized, err := isCanonicalExtension(nEmb, kEmb); err != nil {
			return nil, err
		} else if !canonized {
			// errors.Logf("DEBUG", "%v is not canon (skipping)", sgs[0].Label())
		} else {
			nodes = append(nodes, k)
		}
	}
	return nodes, cache(dt, dt.CanonKidCount, dt.CanonKids, n.Label(), nodes)
}

func children(n *EmbListNode) (nodes []lattice.Node, err error) {
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
	sup, err := dt.Supported(dt, embeddings)
	if err != nil {
		return nil, err
	}
	sizes := set.NewSortedSet(len(embeddings[0].V))
	for _, set := range support.VertexMapSets(embeddings) {
		sizes.Add(types.Int(set.Size()))
	}
	errors.Logf("EMBEDDINGS", "len(V) %v len(embeddings) %v supported %v unique-vertex-embeddings %v", len(embeddings[0].V), len(embeddings), len(sup), sizes)
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
