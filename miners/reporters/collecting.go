package reporters

import ()

import ()

import (
	"github.com/timtadh/sfp/lattice"
)

type Collector struct {
	Nodes []lattice.Node
}

func (c *Collector) Report(n lattice.Node) error {
	c.Nodes = append(c.Nodes, n)
	return nil
}

func (c *Collector) Close() error {
	return nil
}
