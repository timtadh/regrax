package lattice

import (
	"io"
)

type Lattice struct {
	V []Node
	E []Edge
}

type DataType interface {
	Support() int
	Loader() Loader
	Acceptable(Node) bool
	TooLarge(Node) bool
	Close() error
}

type Loader interface {
	StartingPoints(input Input) ([]Node, error)
}

type Input func()(reader io.Reader, closer func())

type Node interface {
	AdjacentCount() (int, error)
	Parents() ([]Node, error)
	ParentCount() (int, error)
	Children() ([]Node, error)
	ChildCount() (int, error)
	Maximal() (bool, error)
	Label() []byte
	Lattice() (*Lattice, error)
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
