package digraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)


type SearchNode struct {
	dt    *Digraph
	pat   *SubGraph
}

func NewSearchNode(dt *Digraph, sg *goiso.SubGraph) *SearchNode {
	return &SearchNode{
		dt: dt,
		pat: NewSubGraph(sg),
	}
}

func LoadSearchNode(dt *Digraph, label []byte) (*SearchNode, error) {
	pat, err := LoadSubgraphFromLabel(label)
	if err != nil {
		return nil, err
	}
	return &SearchNode{dt: dt, pat: pat}, nil
}

func (n *SearchNode) Save() error {
	_, err := n.Embeddings() // ensures that the label is in the embeddings table
	return err
}

func (n *SearchNode) Embeddings() ([]*goiso.SubGraph, error) {
	if has, err := n.dt.Embeddings.Has(n.Label()); err != nil {
		return nil, err
	} else if has {
		return n.loadEmbeddings()
	} else {
		embs, err := n.pat.Embeddings(n.dt)
		if err != nil {
			return nil, err
		}
		err = n.saveEmbeddings(embs)
		if err != nil {
			return nil, err
		}
		return embs, nil
	}
}

func (n *SearchNode) loadEmbeddings() ([]*goiso.SubGraph, error) {
	embs := make([]*goiso.SubGraph, 0, n.dt.Support())
	label := n.Label()
	err := n.dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
		embs = append(embs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return embs, nil
}

func (n *SearchNode) saveEmbeddings(embs []*goiso.SubGraph) (error) {
	label := n.Label()
	for _, emb := range embs {
		err := n.dt.Embeddings.Add(label, emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *SearchNode) loadFrequentVertices() ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, len(n.dt.FrequentVertices))
	for _, label := range n.dt.FrequentVertices {
		node, err := LoadSearchNode(n.dt, label)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (n *SearchNode) Pattern() lattice.Pattern {
	return n
}

func (n *SearchNode) AdjacentCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) Parents() ([]lattice.Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *SearchNode) ParentCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) Children() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "Children of %v", n)
	return n.children(false, n.dt.Children, n.dt.ChildCount)
}

func (n *SearchNode) ChildCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) CanonKids() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return n.children(true, n.dt.CanonKids, n.dt.CanonKidCount)
}

func (n *SearchNode) Maximal() (bool, error) {
	// errors.Logf("DEBUG", "Maximal of %v", n)
	kids, err := n.Children()
	if err != nil {
		return false, err
	}
	return len(kids) == 0, nil
}

func (n *SearchNode) children(checkCanon bool, children bytes_bytes.MultiMap, childCount bytes_int.MultiMap) (nodes []lattice.Node, err error) {
	if len(n.pat.V) == 0 {
		return n.loadFrequentVertices()
	}
	if len(n.pat.E) >= n.dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	/*
	if nodes, has, err := cached(n.dt, childCount, children, n.Label()); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	*/
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) (SubGraphs, error) {
		if n.dt.G.ColorFrequency(e.Color) < n.dt.Support() {
			return exts, nil
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Src].Color) < n.dt.Support() {
			return exts, nil
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Targ].Color) < n.dt.Support() {
			return exts, nil
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			ext, _ := sg.EdgeExtend(e)
			if len(ext.V) > n.dt.MaxVertices {
				return exts, nil
			}
			exts = append(exts, ext)
		}
		return exts, nil
	}
	embs, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	for _, sg := range embs {
		for _, u := range sg.V {
			for _, e := range n.dt.G.Kids[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
			for _, e := range n.dt.G.Parents[u.Id] {
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
		sn := NewSearchNode(n.dt, sgs[0])
		snembs, err := sn.Embeddings()
		if err != nil {
			return nil, err
		}
		if len(snembs) < n.dt.Support() {
			continue
		}
		if len(n.dt.Supported(snembs)) >= n.dt.Support() {
			if checkCanon {
				if canonized, err := isCanonicalExtension(embs[0], sgs[0]); err != nil {
					return nil, err
				} else if !canonized {
					// errors.Logf("DEBUG", "%v is not canon (skipping)", sgs[0].Label())
				} else {
					// errors.Logf("DEBUG", "len(embs) %v len(partition) %v len(supported) %v %v", len(embs), len(sgs), len(n.dt.Supported(embs)), sgs[0].Label())
					nodes = append(nodes, sn)
				}
			} else {
				nodes = append(nodes, sn)
			}
		}
	}
	// errors.Logf("DEBUG", "nodes %v", nodes)
	// return nodes, cache(n.dt, childCount, children, n.Label(), nodes)
	return nodes, nil
}

func (n *SearchNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (n *SearchNode) Level() int {
	return len(n.pat.E) + 1
}

func (n *SearchNode) Label() []byte {
	return n.pat.Label()
}

func (n *SearchNode) String() string {
	sg := n.pat
	V := make([]string, 0, len(sg.V))
	E := make([]string, 0, len(sg.E))
	for _, v := range sg.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			v.Idx,
			n.dt.G.Colors[v.Color],
		))
	}
	for _, e := range sg.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			n.dt.G.Colors[e.Color],
		))
	}
	return fmt.Sprintf("<SearchNode %v:%v%v%v>", len(sg.E), len(sg.V), strings.Join(V, ""), strings.Join(E, ""))
}

func (n *SearchNode) Equals(o types.Equatable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case *Pattern: return a.Equals(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (n *SearchNode) Less(o types.Sortable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case *Pattern: return a.Less(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (n *SearchNode) Hash() int {
	return types.ByteSlice(n.Label()).Hash()
}

