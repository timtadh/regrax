package walker

import (
)

import (
	"github.com/timtadh/data-structures/errors"
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
	Start  []lattice.Node
	Walk   Walk
}

func NewWalker(conf *config.Config, walk Walk) *Walker {
	return &Walker{
		Config: conf,
		Walk: walk,
	}
}

func (w *Walker) Init(dt lattice.DataType) (err error) {
	errors.Logf("INFO", "about to load singleton nodes")
	w.Dt = dt
	w.Start, err = w.Dt.Singletons()
	return err
}

func (w *Walker) Close() error {
	return nil
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	err := w.Init(dt)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "finished initialization, starting walk")
	samples, terminate, errs := w.Walk(w)
	samples = w.RejectingWalk(samples, terminate)
	loop: for {
		select {
		case sampled, open := <-samples:
			if !open {
				break loop
			}
			if err := rptr.Report(sampled); err != nil {
				return err
			}
		case err, open := <-errs:
			if !open {
				break loop
			}
			return err
		}
	}
	return nil
}

func (w *Walker) RejectingWalk(samples chan lattice.Node, terminate chan bool) (chan lattice.Node) {
	nodes := make(chan lattice.Node)
	go func() {
		i := 0
		for sampled := range samples {
			if w.Dt.Acceptable(sampled) {
				nodes<-sampled
				i++
			} else {
				// errors.Logf("INFO", "rejected %v", sampled)
			}
			if i >= w.Config.Samples {
				break
			}
		}
		terminate<-true
		<-samples
		close(nodes)
	}()
	return nodes
}
