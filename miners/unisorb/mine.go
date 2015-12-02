package unisorb

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
	"github.com/timtadh/sfp/stats"
)

func NewWalker(conf *config.Config) *walker.Walker {
	return walker.NewWalker(conf, absorbing.MakeAbsorbingWalk(absorbing.MakeSample(Next), make(chan error)))
}

func Next(cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.Children()
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	if len(kids) <= 0 {
		return nil, nil
	}
	prs, err := transPrs(cur, kids)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return kids[i], nil
}

func transPrs(u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := weight(v)
		if err != nil {
			return nil, err
		}
		weights = append(weights, wght)
		total += wght
	}
	prs := make([]float64, 0, len(adjs))
	for _, wght := range weights {
		prs = append(prs, wght/total)
	}
	return prs, nil
}

func weight(v lattice.Node) (float64, error) {
	vmax, err := v.Maximal()
	if err != nil {
		return 0, err
	}
	vdeg, err := v.ParentCount()
	if err != nil {
		return 0, err
	}
	if vmax {
		return 1.0 / float64(vdeg), nil
	} else {
		return 1.0, nil
	}
}
