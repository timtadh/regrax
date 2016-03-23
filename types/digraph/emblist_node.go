package digraph

import (
	"fmt"
)

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type EmbListNode struct {
	SearchNode
	sgs     []*goiso.SubGraph
}

type Embedding struct {
	sg *goiso.SubGraph
}
func NewEmbListNode(Dt *Digraph, sgs []*goiso.SubGraph) *EmbListNode {
	if len(sgs) > 0 {
		return &EmbListNode{newSearchNode(Dt, sgs[0]), sgs}
	}
	return &EmbListNode{newSearchNode(Dt, nil), nil}
}

func (n *EmbListNode) New(sgs []*goiso.SubGraph) Node {
	return NewEmbListNode(n.Dt, sgs)
}

func LoadEmbListNode(Dt *Digraph, label []byte) (*EmbListNode, error) {
	sgs := make([]*goiso.SubGraph, 0, 10)
	err := Dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewEmbListNode(Dt, sgs), nil
}

func (n *EmbListNode) Pattern() lattice.Pattern {
	return &n.SearchNode
}

func (n *EmbListNode) Embedding() (*goiso.SubGraph, error) {
	// errors.Logf("DEBUG", "Embedding() %v", n)
	if len(n.sgs) == 0 {
		return nil, nil
	} else {
		return n.sgs[0], nil
	}
}

func (n *EmbListNode) Embeddings() ([]*goiso.SubGraph, error) {
	return n.sgs, nil
}

func (n *EmbListNode) Save() error {
	if has, err := n.Dt.Embeddings.Has(n.Label()); err != nil {
		return err
	} else if has {
		return nil
	}
	for _, sg := range n.sgs {
		err := n.Dt.Embeddings.Add(n.Label(), sg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *EmbListNode) String() string {
	if len(n.sgs) > 0 {
		return fmt.Sprintf("<EmbListNode %v>", n.sgs[0].Label())
	} else {
		return fmt.Sprintf("<EmbListNode {}>")
	}
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
	return n.SearchNode.Label()
}

func (n *EmbListNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
