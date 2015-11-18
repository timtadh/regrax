package graph

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
)



type Node struct {
	label []byte
	sgs []*goiso.SubGraph
}

type Embedding struct {
	sg *goiso.SubGraph
}


func (n *Node) Save(dt *Graph) error {
	return errors.Errorf("unimplemented")
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node>")
}

func (n *Node) StartingPoint() bool {
	return n.Size() == 1
}

func (n *Node) Size() int {
	return 0
}

func (n *Node) Parents(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *Node) Children(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *Node) AdjacentCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) ParentCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) ChildCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) Maximal(support int, dtype lattice.DataType) (bool, error) {
	return false, errors.Errorf("unimplemented")
}

func (n *Node) Label() []byte {
	return nil
}

func (n *Node) Embeddings() ([]lattice.Embedding, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *Node) Lattice(support int, dtype lattice.DataType) (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (e *Embedding) Components() ([]int, error) {
	return nil, errors.Errorf("unimplemented")
}

