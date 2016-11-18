package digraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

type SubgraphPattern struct {
	Dt  *Digraph
	Pat *subgraph.SubGraph
}

func NewSubgraphPattern(Dt *Digraph, sg *subgraph.SubGraph) *SubgraphPattern {
	return &SubgraphPattern{
		Dt:  Dt,
		Pat: sg,
	}
}

func LoadSubgraphPattern(Dt *Digraph, label []byte) (*SubgraphPattern, error) {
	pat, err := subgraph.LoadSubGraph(label)
	if err != nil {
		return nil, err
	}
	return NewSubgraphPattern(Dt, pat), nil
}

func (n *SubgraphPattern) dt() *Digraph {
	return n.Dt
}

func (n *SubgraphPattern) SubGraph() *subgraph.SubGraph {
	return n.Pat
}

func (n *SubgraphPattern) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (n *SubgraphPattern) edges() int {
	return len(n.Pat.E)
}

func (n *SubgraphPattern) isRoot() bool {
	return len(n.Pat.E) == 0 && len(n.Pat.V) == 0
}

func (n *SubgraphPattern) Level() int {
	return n.edges() + 1
}

func (n *SubgraphPattern) Label() []byte {
	return n.Pat.Label()
}

func (n *SubgraphPattern) Distance(p lattice.Pattern) float64 {
	o := p.(*SubgraphPattern)
	return n.Pat.Metric(o.Pat)
}

func (n *SubgraphPattern) String() string {
	sg := n.Pat
	V := make([]string, 0, len(sg.V))
	E := make([]string, 0, len(sg.E))
	for _, v := range sg.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			v.Idx,
			n.Dt.Labels.Label(v.Color),
		))
	}
	for _, e := range sg.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			n.Dt.Labels.Label(e.Color),
		))
	}
	return fmt.Sprintf("<SubgraphPattern %v:%v%v%v>", len(sg.E), len(sg.V), strings.Join(V, ""), strings.Join(E, ""))
}

type Labeled interface {
	Label() []byte
}

func (n *SubgraphPattern) Equals(o types.Equatable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case Labeled:
		return a.Equals(types.ByteSlice(b.Label()))
	default:
		return false
	}
}

func (n *SubgraphPattern) Less(o types.Sortable) bool {
	a := types.ByteSlice(n.Label())
	switch b := o.(type) {
	case Labeled:
		return a.Less(types.ByteSlice(b.Label()))
	default:
		return false
	}
}

func (n *SubgraphPattern) Hash() int {
	return types.ByteSlice(n.Label()).Hash()
}
