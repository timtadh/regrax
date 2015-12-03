package reporters

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Collector struct {
	Nodes []lattice.Node
}

func (c *Collector) Report(n lattice.Node) error {
	c.Nodes = append(c.Nodes, n)
	errors.Logf("INFO", "collected %v %v", len(c.Nodes), n)
	return nil
}

func (c *Collector) Close() error {
	return nil
}
