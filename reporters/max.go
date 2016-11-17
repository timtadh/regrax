package reporters

import ()

import ()

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)

type Max struct {
	Reporter miners.Reporter
}

func NewMax(reporter miners.Reporter) (*Max, error) {
	m := &Max{
		Reporter: reporter,
	}
	return m, nil
}

func (r *Max) Report(n lattice.Node) error {
	if ismax, err := n.Maximal(); err != nil {
		return err
	} else if ismax {
		return r.Reporter.Report(n)
	}
	return nil
}

func (r *Max) Close() error {
	return r.Reporter.Close()
}

type CanonMax struct {
	Reporter miners.Reporter
}

func NewCanonMax(reporter miners.Reporter) (*CanonMax, error) {
	m := &CanonMax{
		Reporter: reporter,
	}
	return m, nil
}

func (r *CanonMax) Report(n lattice.Node) error {
	if kids, err := n.CanonKids(); err != nil {
		return err
	} else if len(kids) == 0 {
		return r.Reporter.Report(n)
	}
	return nil
}

func (r *CanonMax) Close() error {
	return r.Reporter.Close()
}
