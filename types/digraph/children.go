package digraph

import (
	"bytes"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	// "github.com/timtadh/sfp/stores/bytes_bytes"
	// "github.com/timtadh/sfp/stores/bytes_int"
)


type Node interface {
	lattice.Node
	New([]*goiso.SubGraph) Node
	Label() []byte
	Embeddings() ([]*goiso.SubGraph, error)
	Embedding() (*goiso.SubGraph, error)
	SubGraph() *SubGraph
	loadFrequentVertices() ([]lattice.Node, error)
	isRoot() bool
	edges() int
	dt() *Digraph
}

func canonChildren(n Node) (nodes []lattice.Node, err error) {
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

func children(n Node) (nodes []lattice.Node, err error) {
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


type SubGraphs []*goiso.SubGraph

func (sgs SubGraphs) Len() int {
	return len(sgs)
}

func (sgs SubGraphs) Less(i, j int) bool {
	return bytes.Compare(sgs[i].ShortLabel(), sgs[j].ShortLabel()) < 0
}

func (sgs SubGraphs) Swap(i, j int) {
	sgs[i], sgs[j] = sgs[j], sgs[i]
}

func (sgs SubGraphs) Verify() error {
	if len(sgs) <= 0 {
		return errors.Errorf("empty partition")
	}
	label := sgs[0].ShortLabel()
	for _, sg := range sgs {
		if !bytes.Equal(label, sg.ShortLabel()) {
			return errors.Errorf("bad partition %v %v", sgs[0].Label(), sg.Label())
		}
	}
	return nil
}

func (sgs SubGraphs) Partition() []SubGraphs {
	sort.Sort(sgs)
	parts := make([]SubGraphs, 0, 10)
	add := func(parts []SubGraphs, buf SubGraphs) []SubGraphs {
		err := buf.Verify()
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		} else {
			parts = append(parts, buf)
		}
		return parts
	}
	buf := make(SubGraphs, 0, 10)
	var ckey []byte = nil
	for _, sg := range sgs {
		label := sg.ShortLabel()
		if ckey != nil && !bytes.Equal(ckey, label) {
			parts = add(parts, buf)
			buf = make(SubGraphs, 0, 10)
		}
		ckey = label
		buf = append(buf, sg)
	}
	if len(buf) > 0 {
		parts = add(parts, buf)
	}
	return parts
}

