package reporters

import (
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)


type ChainReporter struct {
	Reporters []miners.Reporter
}

func (r *ChainReporter) Report(n lattice.Node) error {
	for _, rpt := range r.Reporters {
		err := rpt.Report(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ChainReporter) Close() error {
	for _, rpt := range r.Reporters {
		err := rpt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}



