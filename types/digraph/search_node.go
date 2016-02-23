package digraph

import (
	"fmt"
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

func (n *SearchNode) Pattern() lattice.Pattern {
	return n
}

func (n *SearchNode) AdjacentCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) Parents() ([]Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *SearchNode) ParentCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) Children() ([]Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *SearchNode) ChildCount() (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *SearchNode) CanonKids() ([]Node, error) {
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
	return fmt.Sprintf("<SearchNode %v>", n.pat)
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

