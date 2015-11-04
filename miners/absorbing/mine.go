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

type SparseEntry struct {
	Row, Col int
	Value float64
	Inverse int
}

type Sparse struct {
	Rows, Cols int
	Entries []SparseEntry
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
		Q, R, u, err := m.PrMatrices(sampled, dt)
		if err != nil {
			return err
		}
		errors.Logf("INFO", "matrix Q %v", Q)
		errors.Logf("INFO", "matrix R %v", R)
		errors.Logf("INFO", "matrix u %v", u)
		errors.Logf("INFO", "")
	}
	return nil
}

func (m *Miner) PrMatrices(n lattice.Node, dt lattice.DataType) (Q, R, u *Sparse, err error) {
	lat, err := lattice.MakeLattice(n, m.config.Support, dt)
	if err != nil {
		return nil, nil, nil, err
	}
	p, err := m.probabilities(lat, dt)
	if err != nil {
		return nil, nil, nil, err
	}
	Q = &Sparse{
		Rows: len(lat.V)-1,
		Cols: len(lat.V)-1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	R = &Sparse{
		Rows: len(lat.V)-1,
		Cols: 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	u = &Sparse{
		Rows: 1,
		Cols: len(lat.V)-1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	sp := len(m.start)
	for i, x := range lat.V {
		if x.StartingPoint() && i < len(lat.V)-1 {
			u.Entries = append(u.Entries, SparseEntry{0, i, 1.0/float64(sp), sp})
		}
	}
	for _, e := range lat.E {
		if e.Targ >= len(lat.V)-1 {
			R.Entries = append(R.Entries, SparseEntry{e.Src, 0, 1.0/float64(p[e.Src]), p[e.Src]})
		} else {
			Q.Entries = append(Q.Entries, SparseEntry{e.Src, e.Targ, 1.0/float64(p[e.Src]), p[e.Src]})
		}
	}
	return Q, R, u, nil
}

func (m *Miner) probabilities(lat *lattice.Lattice, dt lattice.DataType) ([]int, error) {
	P := make([]int, len(lat.V))
	for i, node := range lat.V {
		kids, err := node.Children(m.config.Support, dt)
		if err != nil {
			return nil, err
		}
		count := len(kids)
		if i + 1 == len(lat.V) {
			P[i] = -1
		} else if count == 0 {
			return nil, errors.Errorf("0 count for %v", node)
		} else {
			P[i] = count
		}
	}
	return P, nil
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
	next, err := uniform(cur.Children(m.config.Support, dt))
	if err != nil {
		return nil, err
	}
	for next != nil {
		cur = next
		next, err = uniform(cur.Children(m.config.Support, dt))
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

