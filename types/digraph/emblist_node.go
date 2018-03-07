package digraph

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
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
	unsupEmbs  subgraph.VertexEmbeddings
	unsupExts  *set.SortedSet
}

func NewEmbListNode(dt *Digraph, pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding, overlap []map[int]bool, unsupEmbs subgraph.VertexEmbeddings) *EmbListNode {
	if embs != nil {
		if exts == nil {
			panic("nil exts")
		}
		return &EmbListNode{SubgraphPattern{dt, pattern}, exts, embs, overlap, unsupEmbs, nil}
	}
	return &EmbListNode{SubgraphPattern{dt, pattern}, nil, nil, nil, nil, nil}
}

func (n *EmbListNode) New(pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding, overlap []map[int]bool, unsupEmbs subgraph.VertexEmbeddings) Node {
	return NewEmbListNode(n.Dt, pattern, exts, embs, overlap, unsupEmbs)
}

func LoadEmbListNode(dt *Digraph, label []byte) (*EmbListNode, error) {
	sg, err := subgraph.LoadSubGraph(label)
	if err != nil {
		return nil, err
	}
	has, _, exts, embs, overlap, unsupEmbs, err := loadCachedExtsEmbs(dt, sg)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.Errorf("Node was not saved: %v", &SubgraphPattern{Dt: dt, Pat: sg})
	}

	n := &EmbListNode{
		SubgraphPattern: SubgraphPattern{Dt: dt, Pat: sg},
		extensions:      exts,
		embeddings:      embs,
		overlap:         overlap,
		unsupEmbs:       unsupEmbs,
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

func (n *EmbListNode) Overlap() ([]map[int]bool, error) {
	return n.overlap, nil
}

func (n *EmbListNode) UnsupportedExts() (*set.SortedSet, error) {
	if n.unsupExts != nil && n.Dt.Config.Mode&ExtensionPruning == ExtensionPruning {
		return n.unsupExts, nil
	}
	if n.Dt.UnsupExts == nil || n.Dt.Config.Mode&Caching == 0 {
		return set.NewSortedSet(0), nil
	}
	n.Dt.lock.RLock()
	defer n.Dt.lock.RUnlock()
	label := n.Label()
	u := set.NewSortedSet(10)
	err := n.Dt.UnsupExts.DoFind(label, func(_ []byte, ext *subgraph.Extension) error {
		return u.Add(ext)
	})
	if err != nil {
		return nil, err
	}
	n.unsupExts = u
	return u, nil
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
	if n.Dt.UnsupExts == nil || n.Dt.Config.Mode&Caching == 0 {
		return nil
	}
	n.Dt.lock.Lock()
	defer n.Dt.lock.Unlock()
	if len(n.Pat.E) < 4 {
		return nil
	}
	label := n.Label()
	for x, next := n.unsupExts.Items()(); next != nil; x, next = next() {
		ept := x.(*subgraph.Extension)
		err := n.Dt.UnsupExts.Add(label, ept)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *EmbListNode) UnsupportedEmbs() (subgraph.VertexEmbeddings, error) {
	if n.Dt.Config.Mode&EmbeddingPruning == 0 {
		return nil, nil
	}
	return n.unsupEmbs, nil
}

func (n *EmbListNode) String() string {
	return fmt.Sprintf("<EmbListNode %v>", n.Pat.Pretty(n.Dt.Labels))
	//return fmt.Sprintf("<Node %v %v %v>", len(n.embeddings), len(n.extensions), n.Pat.Pretty(n.Dt.G.Colors))
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
