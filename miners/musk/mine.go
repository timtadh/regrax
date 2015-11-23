package musk

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
	samples, errs := m.maxUniformWalk(dt)
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

func (m *Miner) maxUniformWalk(dt lattice.DataType) (chan lattice.Node, chan error) {
	nodes := make(chan lattice.Node)
	errs := make(chan error)
	count := 0
	go func() {
		cur := m.start[rand.Intn(len(m.start))]
		for count < m.config.Samples {
			if ismax, err := cur.Maximal(m.config.Support, dt); err != nil {
				errs <- err
			} else if ismax {
				count++
				nodes <- cur
			}
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

func (m *Miner) transPrs(u lattice.Node, adjs []lattice.Node, dt lattice.DataType) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		w, err := m.weight(u, v, dt)
		if err != nil {
			return nil, err
		}
		weights = append(weights, w)
		total += w
	}
	prs := make([]float64, 0, len(adjs))
	for _, w := range weights {
		prs = append(prs, w/total)
	}
	return prs, nil
}

func (m *Miner) weight(u, v lattice.Node, dt lattice.DataType) (float64, error) {
	umax, err := u.Maximal(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	vmax, err := v.Maximal(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	udeg, err := u.AdjacentCount(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	vdeg, err := v.AdjacentCount(m.config.Support, dt)
	if err != nil {
		return 0, err
	}
	if umax && vmax {
		return 0, nil
	} else if !umax && vmax {
		return 1.0/float64(vdeg), nil
	} else if umax && !vmax {
		return 1.0/float64(udeg), nil
	} else {
		return 1.0, nil
	}
}

