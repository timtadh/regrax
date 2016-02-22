package stack


import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)

type Miner struct {
	Config *config.Config
	Dt     lattice.DataType
	Rptr   miners.Reporter
	Start  []lattice.Node
}

func NewMiner(conf *config.Config) *Miner {
	return &Miner{
		Config: conf,
	}
}

func (m *Miner) PrFormatter() lattice.PrFormatter {
	return nil
}

func (m *Miner) Init(dt lattice.DataType, rptr miners.Reporter) (err error) {
	errors.Logf("INFO", "about to load singleton nodes")
	m.Dt = dt
	m.Rptr = rptr
	m.Start, err = m.Dt.Singletons()
	if err != nil {
		return err
	}
	return nil
}

func (m *Miner) Close() error {
	errors := make(chan error)
	go func() {
		errors <- m.Dt.Close()
	}()
	go func() {
		errors <- m.Rptr.Close()
	}()
	for i := 0; i < 2; i++ {
		err := <-errors
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Miner) Mine(dt lattice.DataType, rptr miners.Reporter, fmtr lattice.Formatter) error {
	err := m.Init(dt, rptr)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "finished initialization, starting walk")
	err = m.mine()
	if err != nil {
		return err
	}
	errors.Logf("INFO", "exiting Mine")
	return nil
}

func (m *Miner) mine() (err error) {
	runes := make([]rune, 0, 10)
	for i := 0; i < 10; i++ {
		runes = append(runes, rune(97 + rand.Intn(26)))
	}
	name := string(runes)
	seen, err := m.Config.BytesIntMultiMap("stack-seen-" + name)
	if err != nil {
		return err
	}
	add := func(stack []lattice.Node, n lattice.Node) ([]lattice.Node, error) {
		err := seen.Add(n.Pattern().Label(), 1)
		if err != nil {
			return nil, err
		}
		return append(stack, n), nil
	}
	pop := func(stack []lattice.Node) ([]lattice.Node, lattice.Node) {
		return stack[:len(stack)-1], stack[len(stack)-1]
	}
	stack := make([]lattice.Node, 0, len(m.Start))
	for _, n := range m.Start {
		stack, err = add(stack, n)
		if err != nil {
			return err
		}
	}
	for len(stack) > 0 {
		var n lattice.Node
		stack, n = pop(stack)
		err = m.Rptr.Report(n)
		if err != nil {
			return err
		}
		kids, err := n.Children()
		if err != nil {
			return err
		}
		for _, k := range kids {
			if has, err := seen.Has(k.Pattern().Label()); err != nil {
				return err
			} else if !has {
				stack, err = add(stack, k)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

