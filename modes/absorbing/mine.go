package absorbing

import (
	"io"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)


type Miner struct {
	Support int
	MinVertices int
	MaxVertices int
	config *config.Config
	start []lattice.Node
}

func NewMiner(conf *config.Config, support int, minVertices int, maxVertices int) *Miner {
	return &Miner{
		Support: support,
		MinVertices: minVertices,
		MaxVertices: maxVertices,
		config: conf,
	}
}

func (m *Miner) Close() error {
	return nil
}

func (m *Miner) Mine(input io.Reader, dt lattice.DataType) error {
	start, err := dt.Loader().StartingPoints(input, m.Support)
	if err != nil {
		return err
	}
	m.start = start
	for _, n := range m.start {
		errors.Logf("INFO", "%v", n)
	}
	panic("unfinished")
}


