package vsigram

import (
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/pool"
)

import (
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/miners"
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

func (m *Miner) mine() (err error) {
	var wg sync.WaitGroup
	pool := pool.New(m.Config.Workers())
	errors.Logf("DEBUG", "pool %v", pool)
	stack := NewStack()
	stack.Push(m.Dt.Root())
	errs := make(chan error)
	reports := make(chan lattice.Node, 100)
	go func() {
		for n := range reports {
			err := m.Rptr.Report(n)
			if err != nil {
				errs<-err
			}
			wg.Done()
		}
	}()
	errList := make([]error, 0, 10)
	go func() {
		for err := range errs {
			if err != nil {
				errList = append(errList, err)
			}
			wg.Done()
		}
	}()
	outer:
	for {
		n := stack.Pop()
		for n == nil {
			i := pool.WaitCount()
			if i <= 0 && stack.Empty() {
				break outer
			}
			n = stack.Pop()
		}
		if n == nil {
			panic("nil")
		}
		err := pool.Do(func(n lattice.Node) func() {
			if n == nil {
				panic("nil")
			}
			return func() {
				if n == nil {
					panic("nil")
				}
				var err error
				err = m.step(&wg, n, reports, stack)
				if err != nil {
					wg.Add(1)
					errs <- err
				}
			}
		}(n))
		if err != nil {
			return err
		}
	}
	pool.Stop()
	close(reports)
	close(errs)
	wg.Wait()
	if len(errList) > 0 {
		return errList[0]
	}
	return nil
}

func (m *Miner) step(wg *sync.WaitGroup, n lattice.Node, reports chan lattice.Node, stack *Stack) (err error) {
	if m.Dt.Acceptable(n) {
		wg.Add(1)
		reports<-n
	}
	kids, err := n.CanonKids()
	if err != nil {
		return err
	}
	for _, k := range kids {
		stack.Push(k)
	}
	return nil
}
