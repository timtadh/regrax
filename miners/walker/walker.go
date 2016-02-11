package walker

import ()

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)

type Walk func(w *Walker) (chan lattice.Node, chan bool, chan error)

type Walker struct {
	Config *config.Config
	Dt     lattice.DataType
	Rptr   miners.Reporter
	Start  []lattice.Node
	Walk   Walk
	Reject bool
	Unique bool
}

func NewWalker(conf *config.Config, walk Walk) *Walker {
	return &Walker{
		Config: conf,
		Walk:   walk,
		Reject: true,
		Unique: true,
	}
}

func (w *Walker) Init(dt lattice.DataType, rptr miners.Reporter) (err error) {
	errors.Logf("INFO", "about to load singleton nodes")
	w.Dt = dt
	w.Rptr = rptr
	w.Start, err = w.Dt.Singletons()
	return err
}

func (w *Walker) Close() error {
	errors := make(chan error)
	go func() {
		errors <- w.Dt.Close()
	}()
	go func() {
		errors <- w.Rptr.Close()
	}()
	for i := 0; i < 2; i++ {
		err := <-errors
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	err := w.Init(dt, rptr)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "finished initialization, starting walk")
	samples, terminate, errs := w.Walk(w)
	samples = w.RejectingWalk(samples, terminate)
loop:
	for {
		select {
		case sampled, open := <-samples:
			errors.Logf("INFO", "got sample %v and open %v", sampled, open)
			if sampled != nil {
				if err := w.Rptr.Report(sampled); err != nil {
					return err
				}
			}
			if !open {
				break loop
			}
		case err := <-errs:
			if err != nil {
				return err
			}
		}
	}
	errors.Logf("INFO", "exiting walker Mine")
	return nil
}

func (w *Walker) RejectingWalk(samples chan lattice.Node, terminate chan bool) chan lattice.Node {
	accepted := make(chan lattice.Node)
	go func() {
		i := 0
		seen := set.NewSortedSet(w.Config.Samples)
		for sampled := range samples {
			accept := false
			if !w.Reject || w.Dt.Acceptable(sampled) {
				label := types.ByteSlice(sampled.Pattern().Label())
				if !w.Unique || !seen.Has(label) {
					if w.Unique {
						seen.Add(label)
					}
					accept = true
					i++
					errors.Logf("INFO", "%v sampled unique %v", seen.Size(), sampled)
				}
			} else {
				errors.Logf("DEBUG", "rejected %v", sampled)
			}
			if i >= w.Config.Samples {
				terminate <- true
			} else {
				terminate <- false
			}
			if accept {
				accepted <- sampled
			}
		}
		close(accepted)
		close(terminate)
	}()
	return accepted
}
