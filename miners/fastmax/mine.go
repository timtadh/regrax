package fastmax

import (
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners/absorbing"
	"github.com/timtadh/sfp/miners/walker"
)

type Walker struct {
	walker.Walker
}

func NewWalker(conf *config.Config) *Walker {
	miner := &Walker{}
	miner.Walker = *walker.NewWalker(conf, absorbing.MakeAbsorbingWalk(absorbing.MakeSample(miner), make(chan error)))
	return miner
}

func (w *Walker) Next(cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.Children()
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	return walker.Transition(cur, kids, w.weight)
}

func (w *Walker) weight(_, v lattice.Node) (float64, error) {
	vmax, err := v.Maximal()
	if err != nil {
		return 0, err
	}
	if vmax {
		indeg, err := v.ParentCount()
		if err != nil {
			return 0, err
		}
		level := float64(v.Pattern().Level())
		maxLevel := float64(w.Dt.LargestLevel())
		return (level)/(float64(indeg)*maxLevel), nil
	} else {
		indeg, err := v.ParentCount()
		if err != nil {
			return 0, err
		}
		odeg, err := v.ChildCount()
		if err != nil {
			return 0, err
		}
		return float64(odeg)/float64(indeg), nil
	}
}
