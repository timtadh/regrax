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
	StartingPoints(input Input, support int) ([]Node, error)
}

type Input func()(reader io.Reader, closer func())

type Node interface {
	Parents(support int, dt DataType) (NodeIterator, error)
	Children(support int, dt DataType) (NodeIterator, error)
	Label() ([]byte, error)
	Embeddings() ([]Embedding, error)
}

type Embedding interface {
	Components() ([]int, error)
}

type SupportMetric interface {
	Supported([]Embedding) ([]Embedding, error)
}

type NodeIterator func()(Node, error, NodeIterator)

func Do(run func(int, DataType) (NodeIterator, error), support int, dt DataType, do func(Node) error) error {
	it, err := run(support, dt)
	if err != nil {
		return err
	}
	var node Node
	for node, err, it = it(); it != nil; node, err, it = it() {
		e := do(node)
		if e != nil {
			return e
		}
	}
	return err
}

func Slice(run func(int, DataType) (NodeIterator, error), support int, dt DataType) ([]Node, error) {
	nodes := make([]Node, 0, 10)
	err := Do(run, support, dt,
		func(n Node) error {
			nodes = append(nodes, n)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

