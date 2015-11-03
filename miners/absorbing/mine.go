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
	config *config.Config
	start []lattice.Node
}

func NewMiner(conf *config.Config) *Miner {
	return &Miner{
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
	for s := 0; s < m.config.Samples; s++ {
		sampled, err := m.rejectingWalk(dt)
		if err != nil {
			return err
		}
		errors.Logf("INFO", "sample %v %v", sampled, sampled.Label())
		parents, err := lattice.Slice(sampled.Parents, m.config.Support, dt)
		if err != nil {
			return err
		}
		errors.Logf("DEBUG", "parents %v", parents)

	}
	return nil
}

func (m *Miner) init(input lattice.Input, dt lattice.DataType) (err error) {
	errors.Logf("INFO", "loading data")
	start, err := dt.Loader().StartingPoints(input, m.config.Support)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "loaded data, about to start mining")
	m.start = start
	return nil
}

func (m *Miner) rejectingWalk(dt lattice.DataType) (max lattice.Node, err error) {
	for {
		sampled, err := m.walk(dt)
		if err != nil {
			return nil, err
		}
		if sampled.Size() >= m.config.MinSize {
			return sampled, nil
		}
	}
}

func (m *Miner) walk(dt lattice.DataType) (max lattice.Node, err error) {
	cur, _ := uniform(m.start, nil)
	next, err := uniform(lattice.Slice(cur.Children, m.config.Support, dt))
	if err != nil {
		return nil, err
	}
	for next != nil {
		cur = next
		next, err = uniform(lattice.Slice(cur.Children, m.config.Support, dt))
		if err != nil {
			return nil, err
		}
		if next != nil && next.Size() > m.config.MaxSize {
			next = nil
		}
	}
	return cur, nil
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

