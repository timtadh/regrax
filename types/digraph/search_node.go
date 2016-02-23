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
)


type SearchNode struct {
	dt    *Graph
	pat   *SubGraph
}

func NewSearchNode(dt *Graph, sg *goiso.SubGraph) *SearchNode {
	return &SearchNode{
		dt: dt,
		pat: NewSubGraph(sg),
	}
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
	return nil, errors.Errorf("unimplemented")
}

func (n *SearchNode) ChildCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) CanonKids() ([]lattice.Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *SearchNode) Maximal() (bool, error) {
	return false, errors.Errorf("unimplemented")
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

