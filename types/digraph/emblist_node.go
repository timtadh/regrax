package digraph

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/set"
)

import (
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/types/digraph/subgraph"
)

type EmbListNode struct {
	SubgraphPattern
	extensions []*subgraph.Extension
	embeddings []*subgraph.Embedding
	overlap    []map[int]bool
	unsupExts  *set.SortedSet
	kids       []lattice.Node
	canonKids  []lattice.Node
	parents    []lattice.Node
}

func NewEmbListNode(dt *Digraph, pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding, overlap []map[int]bool) *EmbListNode {
	if embs != nil {
		if exts == nil {
			panic("nil exts")
		}
		return &EmbListNode{SubgraphPattern{dt, pattern}, exts, embs, overlap, nil, nil, nil, nil}
	}
	return &EmbListNode{SubgraphPattern{dt, pattern}, nil, nil, nil, nil, nil, nil, nil}
}

func (n *EmbListNode) New(pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding, overlap []map[int]bool) Node {
	return NewEmbListNode(n.Dt, pattern, exts, embs, overlap)
}

func (n *EmbListNode) Pattern() lattice.Pattern {
	return &n.SubgraphPattern
}

func (n *EmbListNode) Extensions() ([]*subgraph.Extension, error) {
	return n.extensions, nil
}

func (n *EmbListNode) Embeddings() ([]*subgraph.Embedding, error) {
	return n.embeddings, nil
}

func (n *EmbListNode) Overlap() ([]map[int]bool, error) {
	return n.overlap, nil
}

func (n *EmbListNode) UnsupportedExts() (*set.SortedSet, error) {
	if n.unsupExts != nil && n.Dt.Config.Mode&ExtensionPruning == ExtensionPruning {
		return n.unsupExts, nil
	}
	return set.NewSortedSet(0), nil
}

func (n *EmbListNode) SaveUnsupportedExts(orgLen int, vord []int, eps *set.SortedSet) error {
	if n.Dt.Config.Mode&ExtensionPruning == 0 {
		return nil
	}
	n.unsupExts = set.NewSortedSet(eps.Size())
	for x, next := eps.Items()(); next != nil; x, next = next() {
		ep := x.(*subgraph.Extension)
		ept := ep.Translate(orgLen, vord)
		n.unsupExts.Add(ept)
	}
	return nil
}

func (n *EmbListNode) String() string {
	return fmt.Sprintf("<EmbListNode %v>", n.Pat.Pretty(n.Dt.Labels))
	//return fmt.Sprintf("<Node %v %v %v>", len(n.embeddings), len(n.extensions), n.Pat.Pretty(n.Dt.G.Colors))
}

func (n *EmbListNode) Parents() ([]lattice.Node, error) {
	if n.parents != nil {
		return n.parents, nil
	}
	nodes, err := parents(n)
	if err != nil {
		return nil, err
	}
	n.parents = nodes
	return nodes, nil
}

func (n *EmbListNode) Children() (nodes []lattice.Node, err error) {
	if n.kids != nil {
		return n.kids, nil
	}
	nodes, err = children(n)
	if err != nil {
		return nil, err
	}
	n.kids = nodes
	return nodes, nil
}

func (n *EmbListNode) CanonKids() (nodes []lattice.Node, err error) {
	if n.canonKids != nil {
		return n.canonKids, nil
	}
	nodes, err = canonChildren(n)
	if err != nil {
		return nil, err
	}
	n.canonKids = nodes
	return nodes, nil
}

func (n *EmbListNode) loadFrequentVertices() ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, len(n.Dt.FrequentVertices))
	for _, node := range n.Dt.FrequentVertices {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (n *EmbListNode) AdjacentCount() (int, error) {
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

func (n *EmbListNode) ParentCount() (int, error) {
	parents, err := n.Parents()
	if err != nil {
		return 0, err
	}
	return len(parents), nil
}

func (n *EmbListNode) ChildCount() (int, error) {
	kids, err := n.Children()
	if err != nil {
		return 0, err
	}
	return len(kids), nil
}

func (n *EmbListNode) Maximal() (bool, error) {
	cc, err := n.ChildCount()
	if err != nil {
		return false, err
	}
	return cc == 0, nil
}

func (n *EmbListNode) Label() []byte {
	return n.SubgraphPattern.Label()
}

func (n *EmbListNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
