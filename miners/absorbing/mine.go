package absorbing

import (
	"encoding/binary"
	"math/rand"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

func init() {
	if urandom, err := os.Open("/dev/urandom"); err != nil {
		panic(err)
	} else {
		seed := make([]byte, 8)
		if _, err := urandom.Read(seed); err == nil {
			rand.Seed(int64(binary.BigEndian.Uint64(seed)))
		}
		urandom.Close()
	}
}


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

func (m *Miner) Mine(input lattice.Input, dt lattice.DataType) error {
	err := m.init(input, dt)
	if err != nil {
		return err
	}
	path, err := m.walk(dt)
	if err != nil {
		return err
	}
	for i, n := range path {
		errors.Logf("INFO", "%d %v", i, n)
	}
	panic("unfinished")
}

func (m *Miner) init(input lattice.Input, dt lattice.DataType) (err error) {
	errors.Logf("INFO", "loading data")
	start, err := dt.Loader().StartingPoints(input, m.Support)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "loaded data, about to start mining")
	m.start = start
	return nil
}

func (m *Miner) walk(dt lattice.DataType) (path []lattice.Node, err error) {
	cur, _ := uniform(m.start, nil)
	next, err := uniform(lattice.Slice(cur.Children, m.Support, dt))
	if err != nil {
		return nil, err
	}
	path = append(path, cur)
	for next != nil {
		cur = next
		next, err = uniform(lattice.Slice(cur.Children, m.Support, dt))
		if err != nil {
			return nil, err
		}
		path = append(path, cur)
	}
	return path, nil
}

func uniform(slice []lattice.Node, err error) (lattice.Node, error) {
	if err != nil {
		return nil, err
	}
	if len(slice) > 0 {
		return slice[rand.Intn(len(slice))], nil
	}
	return nil, nil
}

