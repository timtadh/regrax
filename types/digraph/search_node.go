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


type SearchNode struct {
	Dt    *Digraph
	Pat   *subgraph.SubGraph
}

func newSearchNode(Dt *Digraph, sg *goiso.SubGraph) SearchNode {
	return SearchNode{
		Dt: Dt,
		Pat: subgraph.NewSubGraph(sg),
	}
}

func NewSearchNode(Dt *Digraph, sg *goiso.SubGraph) *SearchNode {
	n := newSearchNode(Dt, sg)
	return &n
}

func (n *SearchNode) New(sgs []*goiso.SubGraph) Node {
	if len(sgs) > 0 {
		return NewSearchNode(n.Dt, sgs[0])
	}
	return NewSearchNode(n.Dt, nil)
}

func LoadSearchNode(Dt *Digraph, label []byte) (*SearchNode, error) {
	pat, err := subgraph.LoadSubgraphFromLabel(label)
	if err != nil {
		return nil, err
	}
	return &SearchNode{Dt: Dt, Pat: pat}, nil
}

func (n *SearchNode) dt() *Digraph {
	return n.Dt
}

func (n *SearchNode) Save() error {
	_, err := n.Embeddings() // ensures that the label is in the embeddings table
	return err
}

func (n *SearchNode) SubGraph() *subgraph.SubGraph {
	return n.Pat
}

func (n *SearchNode) Embedding() (*goiso.SubGraph, error) {
	embs, err := n.Embeddings()
	if err != nil {
		return nil, err
	} else if len(embs) == 0 {
		return nil, nil
	}
	return embs[0], nil
}

func (n *SearchNode) Embeddings() ([]*goiso.SubGraph, error) {
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
	pc, err := n.ParentCount()
	if err != nil {
		return 0, err
	}
	cc, err := n.ChildCount()
	if err != nil {
		return 0, err
	}
	return pc + cc, nil
}

func (n *SearchNode) Parents() ([]lattice.Node, error) {
	return parents(n, n.Dt.Parents, n.Dt.ParentCount)
}

func (n *SearchNode) ParentCount() (int, error) {
	return count(n, n.Parents, n.Dt.ParentCount)
}

func (n *SearchNode) Children() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "Children of %v", n)
	return children(n)
}

func (n *SearchNode) ChildCount() (int, error) {
	return count(n, n.Children, n.Dt.ChildCount)
}

func (n *SearchNode) CanonKids() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return canonChildren(n)
}

func (n *SearchNode) Maximal() (bool, error) {
	count, err := n.ChildCount()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (n *SearchNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (n *SearchNode) edges() int {
	return len(n.Pat.E)
}

func (n *SearchNode) isRoot() bool {
	return len(n.Pat.E) == 0 && len(n.Pat.V) == 0
}

func (n *SearchNode) Level() int {
	return n.edges() + 1
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

