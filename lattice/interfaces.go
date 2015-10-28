package lattice

import (
	"io"
)

type DataType interface {
	Metric() SupportMetric
	Loader() Loader
	Close() error
}

type Loader interface {
	StartingPoints(input io.Reader, support int) ([]Node, error)
}

type Node interface {
	Parents(support int, metric SupportMetric) (NodeIterator, error)
	Children(support int, metric SupportMetric) (NodeIterator, error)
	Label() ([]byte, error)
	Embeddings() ([]Embedding, error)
}

type Embedding interface {
	Components() ([]int, error)
}

type SupportMetric interface {
	Supported([]Embedding) ([]Embedding, error)
}

type NodeIterator func()(Node, NodeIterator, error)

