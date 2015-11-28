package walker

import (
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
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

func (w *Walker) Init(input lattice.Input, dt lattice.DataType) error {
	errors.Logf("INFO", "loading data")
	start, err := dt.Loader().StartingPoints(input, w.Config.Support)
	if err != nil {
		return err
	}
	errors.Logf("INFO", "loaded data, about to start mining")
	w.Start = start
	w.Dt = dt
	return nil
}

func (w *Walker) Close() error {
	return nil
}

func (w *Walker) Mine(input lattice.Input, dt lattice.DataType) error {
	err := w.Init(input, dt)
	if err != nil {
		return err
	}
	samples, terminate, errs := w.Walk(w)
	samples = w.RejectingWalk(samples, terminate)
	loop: for {
		select {
		case sampled, open := <-samples:
			if !open {
				break loop
			}
			errors.Logf("INFO", "sample %v", sampled)
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
				errors.Logf("INFO", "rejected %v", sampled)
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
