package lattice

import (
	"io"
)

type Lattice struct {
	V []Node
	E []Edge
}

type DataType interface {
	Metric() SupportMetric
	Loader() Loader
	Close() error
}

type Loader interface {
	StartingPoints(input Input, support int) ([]Node, error)
}

type Input func()(reader io.Reader, closer func())

type Node interface {
	StartingPoint() bool
	AdjacentCount(support int, dt DataType) (int, error)
	Parents(support int, dt DataType) ([]Node, error)
	ParentCount(support int, dt DataType) (int, error)
	Children(support int, dt DataType) ([]Node, error)
	ChildCount(support int, dt DataType) (int, error)
	Maximal(support int, dt DataType) (bool, error)
	Size() int
	Label() []byte
	Embeddings() ([]Embedding, error)
	Lattice(support int, dt DataType) (*Lattice, error)
}

type NoLattice struct{}

func (n *NoLattice) Error() string {
	return "No Lattice Function Implemented"
}

type Edge struct {
	Src, Targ int
}

type Embedding interface {
	Components() ([]int, error)
}

type SupportMetric interface {
	Supported([]Embedding) ([]Embedding, error)
}

type NodeIterator func()(Node, error, NodeIterator)
