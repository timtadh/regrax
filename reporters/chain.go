package reporters

import ()

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
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
