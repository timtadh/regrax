package absorbing

import (
	"io"
)

import (
	"github.com/timtadh/sfp/lattice"
)


type Miner struct {
	Support int
	MinVertices int
	MaxVertices int
}

func NewMiner(support int, minVertices int, maxVertices int) *Miner {
	return &Miner{
		Support: support,
		MinVertices: minVertices,
		MaxVertices: maxVertices,
	}
}

func (m *Miner) Mine(input io.Reader, dt lattice.DataType) {
	panic("unimplemented")
}

