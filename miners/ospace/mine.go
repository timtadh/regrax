package ospace

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

func (m *Miner) Close() error {
	return nil
}

func (m *Miner) Mine(input lattice.Input, dt lattice.DataType) error {
	err := m.init(input, dt)
	if err != nil {
		return err
	}
	samples, errs := m.uniformWalk(dt)
	go func() {
		for sampled := range samples {
			errors.Logf("INFO", "sample %v", sampled)
		}
	}()
	for err := range errs {
		return err
	}
	return nil
}

func (m *Miner) uniformWalk(dt lattice.DataType) (chan lattice.Node, chan error) {
	nodes := make(chan lattice.Node)
	errs := make(chan error)
	count := 0
	go func() {
		cur := m.start[rand.Intn(len(m.start))]
		for count < m.config.Samples {
			count++
			nodes <- cur
			next, err := m.next(cur, dt)
			if err != nil {
				errs <- err
				break
			}
			if next == nil {
				errs <- errors.Errorf("next was nil!!")
				break
			}
			cur = next
		}
		close(nodes)
		close(errs)
	}()
	return nodes, errs
}

func (m *Miner) next(cur lattice.Node, dt lattice.DataType) (lattice.Node, error) {
	kids, err := cur.Children(m.config.Support, dt)
	if err != nil {
		return nil, err
	}
	parents, err := cur.Parents(m.config.Support, dt)
	if err != nil {
		return nil, err
	}
	adjs := append(kids, parents...)
	errors.Logf("DEBUG", "cur %v parents %v kids %v adjs %v", cur, len(parents), len(kids), len(adjs))
	prs, err := m.transPrs(cur, adjs, dt)
	if err != nil {
		return nil, err
	}
	adjs = append(adjs, cur)
	prs = append(prs, m.selfPr(prs))
	i := sample(prs)
	return adjs[i], nil
}

func sample(prs []float64) int {
	total := sum(prs)
	i := 0
	x := total * (1 - rand.Float64())
	for x > prs[i] {
		x -= prs[i]
		i += 1
	}
	return i
}

func sum(list []float64) float64 {
	var sum float64
	for _, item := range list {
		sum += item
	}
	return sum
}

func (m *Miner) selfPr(prs []float64) float64 {
	return 1.0 - sum(prs)
}

func (m *Miner) transPrs(u lattice.Node, adjs []lattice.Node, dt lattice.DataType) ([]float64, error) {
	prs := make([]float64, 0, len(adjs))
	for _, v := range adjs {
		w, err := m.weight(u, v, dt)
		if err != nil {
			return nil, err
		}
		// errors.Logf("DEBUG", "u %v, v %v, weight: %v", u, v, w)
		prs = append(prs, 1.0/w)
	}
	return prs, nil
}

func (m *Miner) weight(u, v lattice.Node, dt lattice.DataType) (float64, error) {
	udeg, err := u.AdjacentCount(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	vdeg, err := v.AdjacentCount(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	return max(float64(udeg), float64(vdeg)), nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

