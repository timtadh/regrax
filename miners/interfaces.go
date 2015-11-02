package miners

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)



type Miner interface {
	Mine(lattice.Input, lattice.DataType) error
	Close() error
}

