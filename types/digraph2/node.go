package digraph2

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)

type Node struct {
	dt *Digraph
	SubGraph *subgraph.SubGraph
	Embeddings []*subgraph.Embedding
}

func NewNode(dt *Digraph, sg *subgraph.SubGraph, embs []*subgraph.Embedding) *Node {
	return &Node{
		dt: dt,
		SubGraph: sg,
		Embeddings: embs,
	}
}

func (n *Node) AsNode() lattice.Node {
	return n
}

func (n *Node) Pattern() lattice.Pattern {
	return &Pattern{*n.SubGraph}
}

func (n *Node) String() string {
	if n.SubGraph == nil {
		return "<Node {0:0}>"
	}
	return fmt.Sprintf("<Node %v>", n.SubGraph.Pretty(n.dt.Labels))
}

func (n *Node) Parents() ([]lattice.Node, error) {
	return nil, errors.Errorf("not supported yet")
}

func (n *Node) Children() (nodes []lattice.Node, err error) {
	return n.findChildren(nil)
}

func (n *Node) CanonKids() (nodes []lattice.Node, err error) {
	return n.findChildren(func(ext *subgraph.SubGraph) (bool, error) {
		return isCanonicalExtension(n.SubGraph, ext)
	})
}

func (n *Node) AdjacentCount() (int, error) {
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

func (n *Node) ParentCount() (int, error) {
	return 0, errors.Errorf("not supported yet")
}

func (n *Node) ChildCount() (int, error) {
	return 0, errors.Errorf("not supported yet")
}

func (n *Node) Maximal() (bool, error) {
	cc, err := n.ChildCount()
	if err != nil {
		return false, err
	}
	return cc == 0, nil
}

func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
