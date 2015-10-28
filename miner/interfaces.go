package miner

import (
	"io"
)

import (
	"github.com/timtadh/sfp/lattice"
)


type Miner interface {
	Mine(io.Reader, lattice.DataType) error
	Close() error
}

