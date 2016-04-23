package digraph

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

type EmbListNode struct {
	SubgraphPattern
	extensions []*subgraph.Extension
	embeddings []*subgraph.Embedding
}

type Embedding struct {
	sg *goiso.SubGraph
}

func NewEmbListNode(dt *Digraph, pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding) *EmbListNode {
	if len(embs) > 0 {
		if exts == nil {
			panic("nil exts")
		}
		return &EmbListNode{SubgraphPattern{dt, pattern}, exts, embs}
	}
		return &EmbListNode{SubgraphPattern{dt, pattern}, nil, nil}
}

func (n *EmbListNode) New(pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding) Node {
	return NewEmbListNode(n.Dt, pattern, exts, embs)
}

func LoadEmbListNode(dt *Digraph, label []byte) (*EmbListNode, error) {
	sg, err := subgraph.LoadSubGraph(label)
	if err != nil {
		return nil, err
	}
	has, exts, embs, err := loadCachedExtsEmbs(dt, sg)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.Errorf("Node was not saved: %v", &SubgraphPattern{Dt: dt, Pat: sg})
	}

	n := &EmbListNode{
		SubgraphPattern: SubgraphPattern{Dt: dt, Pat: sg},
		extensions: exts,
		embeddings: embs,
	}
	return n, nil
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

func (n *EmbListNode) String() string {
	return fmt.Sprintf("<EmbListNode %v>", n.Pat.Pretty(n.Dt.G.Colors))
}

func (n *EmbListNode) Parents() ([]lattice.Node, error) {
	return parents(n, n.Dt.Parents, n.Dt.ParentCount)
}

func (n *EmbListNode) Children() (nodes []lattice.Node, err error) {
	return children(n)
}

func (n *EmbListNode) CanonKids() (nodes []lattice.Node, err error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return canonChildren(n)
}

func (n *EmbListNode) loadFrequentVertices() ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, len(n.Dt.FrequentVertices))
	for _, label := range n.Dt.FrequentVertices {
		node, err := LoadEmbListNode(n.Dt, label)
		if err != nil {
			return nil, err
		}
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
	return count(n, n.Parents, n.Dt.ParentCount)
}

func (n *EmbListNode) ChildCount() (int, error) {
	return count(n, n.Children, n.Dt.ChildCount)
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
