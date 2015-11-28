package miners

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)



type Miner interface {
	Mine(lattice.Input, lattice.DataType, Reporter) error
	Close() error
}

type Reporter interface {
	Report(lattice.Node) error
	Close() error
}

