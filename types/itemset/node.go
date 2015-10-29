package itemset

import (
	"github.com/timtadh/sfp/lattice"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
)


type Node struct {
	items *set.SortedSet
	embeddings []int32
}

func (n *Node) Parents(support int, dtype lattice.DataType) (lattice.NodeIterator, error) {
	panic("unimplemented")
}

func (n *Node) Children(support int, dtype lattice.DataType) (lattice.NodeIterator, error) {
	dt := dtype.(*ItemSets)
	errors.Logf("INFO", "dt %v", dt)
	panic("unimplemented")
}

func (n *Node) Label() ([]byte, error) {
	panic("unimplemented")
}

func (n *Node) Embeddings() ([]lattice.Embedding, error) {
	panic("unimplemented")
}

