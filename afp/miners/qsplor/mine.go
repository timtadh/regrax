package qsplor

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/stats"
)


type Miner struct {
	Config       *config.Config
	Dt           lattice.DataType
	Rptr         miners.Reporter
	Scorer       Scorer
	MaxQueueSize int
}

func NewMiner(conf *config.Config, scorer Scorer, maxQueueSize int) *Miner {
	return &Miner{
		Config: conf,
		Scorer: scorer,
		MaxQueueSize: maxQueueSize,
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

func (m *Miner) mine() (err error) {
	seen, err := m.Config.BytesIntMultiMap("stack-seen")
	if err != nil {
		return err
	}
	add := func(stack []lattice.Node, n lattice.Node) ([]lattice.Node, error) {
		err := seen.Add(n.Pattern().Label(), 1)
		if err != nil {
			return nil, err
		}
		stack = append(stack, n)
		if len(stack) > m.MaxQueueSize {
			stack = m.dropOne(stack)
		}
		return stack, nil
	}
	root := m.Dt.Root()
	rootKids, err := root.Children()
	if err != nil {
		return err
	}
	for _, rk := range rootKids {
		stack := make([]lattice.Node, 0, 10)
		stack, err = add(stack, rk)
		if err != nil {
			return err
		}
		for len(stack) > 0 {
			var n lattice.Node
			stack, n = m.takeOne(stack)
			if m.Dt.Acceptable(n) {
				err = m.Rptr.Report(n)
				if err != nil {
					return err
				}
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
	}
	return nil
}

func (m *Miner) takeOne(queue []lattice.Node) ([]lattice.Node, lattice.Node) {
	s := stats.Sample(10, len(queue))
	k := m.Scorer.Kernel(queue, s)
	var i int
	var ms float64
	if len(k) > 0 {
		i, ms = stats.Max(stats.Srange(len(k)), func(i int) float64 { return k.Mean(i) })
		i = s[i]
	} else {
		i, ms = stats.Max(s, func(i int) float64 { return m.Scorer.Score(queue[i], queue) })
	}
	errors.Logf("DEBUG", "max score %v, queue len %v, taking %v", ms, len(queue), queue[i])
	return pop(queue, i)
}

func (m *Miner) dropOne(queue []lattice.Node) []lattice.Node {
	s := stats.Sample(10, len(queue))
	k := m.Scorer.Kernel(queue, s)
	var i int
	var ms float64
	if len(k) > 0 {
		i, ms = stats.Min(stats.Srange(len(k)), func(i int) float64 { return k.Mean(i) })
		i = s[i]
	} else {
		i, ms = stats.Min(s, func(i int) float64 { return m.Scorer.Score(queue[i], queue) })
	}
	errors.Logf("DEBUG", "min score %v, queue len %v, dropping %v", ms, len(queue), queue[i])
	queue, _ = pop(queue, i)
	return queue
}

func pop(stack []lattice.Node, i int) ([]lattice.Node, lattice.Node) {
	item := stack[i]
	if i < len(stack) - 1 {
		copy(stack[i:len(stack)-1], stack[i+1:len(stack)])
	}
	return stack[:len(stack)-1], item
}

