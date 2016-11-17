package reporters

import ()

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)

type Skip struct {
	Skip     int
	Reporter miners.Reporter
	count    int
}

func NewSkip(n int, rptr miners.Reporter) *Skip {
	return &Skip{
		Skip: n,
		Reporter: rptr,
	}
}

func (r *Skip) Report(n lattice.Node) error {
	r.count++
	if r.count%r.Skip == 0 {
		return r.Reporter.Report(n)
	}
	return nil
}

func (r *Skip) Close() error {
	return r.Reporter.Close()
}
