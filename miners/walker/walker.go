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


type Walk func(w *Walker) (chan lattice.Node, chan error)

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
	samples, errs := w.Walk(w)
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
