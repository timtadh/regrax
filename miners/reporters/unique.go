package reporters

import (
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)

type Unique struct {
	Seen *set.SortedSet
	Reporter miners.Reporter
}

func NewUnique(reporter miners.Reporter) *Unique {
	return &Unique{
		Seen: set.NewSortedSet(10),
		Reporter: reporter,
	}
}

func (r *Unique) Report(n lattice.Node) error {
	label := types.ByteSlice(n.Label())
	if r.Seen.Has(label) {
		return nil
	}
	r.Seen.Add(label)
	return r.Reporter.Report(n)
}

func (r *Unique) Close() error {
	return r.Reporter.Close()
}

