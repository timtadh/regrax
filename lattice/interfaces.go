package lattice

import (
	"io"
)

import (
	"github.com/timtadh/data-structures/types"
)

type Lattice struct {
	V []Node
	E []Edge
}

type Input func() (reader io.Reader, closer func())

type Loader interface {
	Load(input Input) (DataType, error)
}

type DataType interface {
	LargestLevel() int
	MinimumLevel() int
	Support() int
	Acceptable(Node) bool
	TooLarge(Node) bool
	Empty() Node
	Singletons() ([]Node, error)
	Close() error
}

type Node interface {
	Pattern() Pattern
	AdjacentCount() (int, error)
	Parents() ([]Node, error)
	ParentCount() (int, error)
	Children() ([]Node, error)
	ChildCount() (int, error)
	CanonKids() ([]Node, error)
	Maximal() (bool, error)
	Lattice() (*Lattice, error)
}

type Pattern interface {
	types.Hashable
	Label() []byte
	Level() int
}

type Formatter interface {
	FileExt() string
	PatternName(Node) string
	Pattern(Node) (string, error)
	Embeddings(Node) ([]string, error)
	FormatPattern(io.Writer, Node) error
	FormatEmbeddings(io.Writer, Node) error
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

type NodeIterator func() (Node, error, NodeIterator)
