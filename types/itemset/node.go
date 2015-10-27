package itemset

import (
	"github.com/timtadh/sfp/lattice"
)


type Node struct {
	items []int
	embeddings []int
}

func (n *Node) Parents(support int, metric lattice.SupportMetric) lattice.NodeIterator {
	panic("unimplemented")
}

func (n *Node) Children(support int, metric lattice.SupportMetric) lattice.NodeIterator {
	panic("unimplemented")
}

func (n *Node) Label() []byte {
	panic("unimplemented")
}

func (n *Node) Embeddings() []lattice.Embedding {
	panic("unimplemented")
}

