package digraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

type SubgraphPattern struct {
	Dt  *Digraph
	Pat *subgraph.SubGraph
}

func newSubgraphPattern(Dt *Digraph, sg *goiso.SubGraph) SubgraphPattern {
	return SubgraphPattern{
		Dt:  Dt,
		Pat: subgraph.FromEmbedding(sg),
	}
}

func NewSubgraphPattern(Dt *Digraph, sg *goiso.SubGraph) *SubgraphPattern {
	n := newSubgraphPattern(Dt, sg)
	return &n
}

func LoadSubgraphPattern(Dt *Digraph, label []byte) (*SubgraphPattern, error) {
	pat, err := subgraph.FromLabel(label)
	if err != nil {
		return nil, err
	}
	return &SubgraphPattern{Dt: Dt, Pat: pat}, nil
}

func (n *SubgraphPattern) dt() *Digraph {
	return n.Dt
}

func (n *SubgraphPattern) Save() error {
	_, err := n.Embeddings() // ensures that the label is in the embeddings table
	return err
}

func (n *SubgraphPattern) SubGraph() *subgraph.SubGraph {
	return n.Pat
}

func (n *SubgraphPattern) Embedding() (*goiso.SubGraph, error) {
	embs, err := n.Embeddings()
	if err != nil {
		return nil, err
	} else if len(embs) == 0 {
		return nil, nil
	}
	return embs[0], nil
}

func (n *SubgraphPattern) Embeddings() ([]*goiso.SubGraph, error) {
	if has, err := n.Dt.Embeddings.Has(n.Label()); err != nil {
		return nil, err
	} else if has {
		return n.loadEmbeddings()
	} else {
		embs, err := n.Pat.Embeddings(n.Dt.G, n.Dt.ColorMap, n.Dt.Extender)
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

func (n *SubgraphPattern) loadEmbeddings() ([]*goiso.SubGraph, error) {
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

func (n *SubgraphPattern) saveEmbeddings(embs []*goiso.SubGraph) error {
	label := n.Label()
	for _, emb := range embs {
		err := n.Dt.Embeddings.Add(label, emb)
		if err != nil {
			return err
		}
	}
	return nil
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

func (n *SubgraphPattern) String() string {
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
