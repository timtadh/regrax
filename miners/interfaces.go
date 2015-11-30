package miners

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)


// Note: the miner's Close function should close both reporter and the datatype that were passed into it.
type Miner interface {
	Mine(lattice.DataType, Reporter) error
	Close() error
}

type Reporter interface {
	Report(lattice.Node) error
	Close() error
}

