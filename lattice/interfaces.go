package lattice

import (
	"io"
)

type DataType interface {
	Metric() SupportMetric
	Loader() Loader
}

type Loader interface {
	StartingPoints(input io.Reader, support int) []Node
}

type Node interface {
	Parents(support int, metric SupportMetric) NodeIterator
	Children(support int, metric SupportMetric) NodeIterator
	Label() []byte
	Embeddings() []Embedding
}

type Embedding interface {
	Components() []int
}

type SupportMetric interface {
	Supported([]Embedding) []Embedding
}

type NodeIterator func()(Node, NodeIterator)

