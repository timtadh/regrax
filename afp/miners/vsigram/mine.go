package vsigram


import ()

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

func (m *Miner) mine() error {
	pop := func(stack []lattice.Node) ([]lattice.Node, lattice.Node) {
		return stack[:len(stack)-1], stack[len(stack)-1]
	}
	stack := make([]lattice.Node, 0, 10)
	stack = append(stack, m.Dt.Root())
	for len(stack) > 0 {
		var n lattice.Node
		stack, n = pop(stack)
		err := m.Rptr.Report(n)
		if err != nil {
			return err
		}
		kids, err := n.CanonKids()
		if err != nil {
			return err
		}
		for _, k := range kids {
			stack = append(stack, k)
		}
	}
	return nil
}

