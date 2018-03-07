package reporters

import ()

import (
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/miners"
)

type Chain struct {
	Reporters []miners.Reporter
}

func (r *Chain) Report(n lattice.Node) error {
	for _, rpt := range r.Reporters {
		err := rpt.Report(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Chain) Close() error {
	for _, rpt := range r.Reporters {
		err := rpt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
