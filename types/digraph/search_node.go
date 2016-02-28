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
	Dt    *Digraph
	Pat   *SubGraph
}

func newSearchNode(Dt *Digraph, sg *goiso.SubGraph) SearchNode {
	return SearchNode{
		Dt: Dt,
		Pat: NewSubGraph(sg),
	}
}

func NewSearchNode(Dt *Digraph, sg *goiso.SubGraph) *SearchNode {
	n := newSearchNode(Dt, sg)
	return &n
}

func LoadSearchNode(Dt *Digraph, label []byte) (*SearchNode, error) {
	pat, err := LoadSubgraphFromLabel(label)
	if err != nil {
		return nil, err
	}
	return &SearchNode{Dt: Dt, Pat: pat}, nil
}

func (n *SearchNode) Save() error {
	_, err := n.Embeddings() // ensures that the label is in the embeddings table
	return err
}

func (n *SearchNode) Embeddings() ([]*goiso.SubGraph, error) {
	if has, err := n.Dt.Embeddings.Has(n.Label()); err != nil {
		return nil, err
	} else if has {
		return n.loadEmbeddings()
	} else {
		embs, err := n.Pat.Embeddings(n.Dt)
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
	embs := make([]*goiso.SubGraph, 0, n.Dt.Support())
	label := n.Label()
	err := n.Dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
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
		err := n.Dt.Embeddings.Add(label, emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *SearchNode) loadFrequentVertices() ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, len(n.Dt.FrequentVertices))
	for _, label := range n.Dt.FrequentVertices {
		node, err := LoadSearchNode(n.Dt, label)
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
	return n.children(false, n.Dt.Children, n.Dt.ChildCount)
}

func (n *SearchNode) ChildCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) CanonKids() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return n.children(true, n.Dt.CanonKids, n.Dt.CanonKidCount)
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
	if len(n.Pat.V) == 0 {
		return n.loadFrequentVertices()
	}
	if len(n.Pat.E) >= n.Dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := cached(n.Dt, childCount, children, n.Label()); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) (SubGraphs, error) {
		if n.Dt.G.ColorFrequency(e.Color) < n.Dt.Support() {
			return exts, nil
		} else if n.Dt.G.ColorFrequency(n.Dt.G.V[e.Src].Color) < n.Dt.Support() {
			return exts, nil
		} else if n.Dt.G.ColorFrequency(n.Dt.G.V[e.Targ].Color) < n.Dt.Support() {
			return exts, nil
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			ext, _ := sg.EdgeExtend(e)
			if len(ext.V) > n.Dt.MaxVertices {
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
			for _, e := range n.Dt.G.Kids[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
			for _, e := range n.Dt.G.Parents[u.Id] {
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
		sn := NewSearchNode(n.Dt, sgs[0])
		snembs, err := sn.Embeddings()
		if err != nil {
			return nil, err
		}
		if len(snembs) < n.Dt.Support() {
			continue
		}
		if len(n.Dt.Supported(snembs)) >= n.Dt.Support() {
			if checkCanon {
				if canonized, err := isCanonicalExtension(embs[0], sgs[0]); err != nil {
					return nil, err
				} else if !canonized {
					// errors.Logf("DEBUG", "%v is not canon (skipping)", sgs[0].Label())
				} else {
					// errors.Logf("DEBUG", "len(embs) %v len(partition) %v len(supported) %v %v", len(embs), len(sgs), len(n.Dt.Supported(embs)), sgs[0].Label())
					nodes = append(nodes, sn)
				}
			} else {
				nodes = append(nodes, sn)
			}
		}
	}
	// errors.Logf("DEBUG", "nodes %v", nodes)
	return nodes, cache(n.Dt, childCount, children, n.Label(), nodes)
}

func (n *SearchNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (n *SearchNode) Level() int {
	return len(n.Pat.E) + 1
}

func (n *SearchNode) Label() []byte {
	return n.Pat.Label()
}

func (n *SearchNode) String() string {
	sg := n.Pat
	V := make([]string, 0, len(sg.V))
	E := make([]string, 0, len(sg.E))
	for _, v := range sg.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			v.Idx,
			n.Dt.G.Colors[v.Color],
		))
	}
	for _, e := range sg.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			n.Dt.G.Colors[e.Color],
		))
	}
	return fmt.Sprintf("<SearchNode %v:%v%v%v>", len(sg.E), len(sg.V), strings.Join(V, ""), strings.Join(E, ""))
}

type Labeled interface {
	Label() []byte
}

func (n *SearchNode) Equals(o types.Equatable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case Labeled: return a.Equals(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (n *SearchNode) Less(o types.Sortable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case Labeled: return a.Less(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (n *SearchNode) Hash() int {
	return types.ByteSlice(n.Label()).Hash()
}

