package premusk

import (
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	// "github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/musk"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/stats"
)


type Walker struct {
	walker.Walker
	Teleports *hashtable.LinearHash
}

func NewWalker(conf *config.Config) *Walker {
	miner := &Walker{
		Teleports: hashtable.NewLinearHash(),
	}
	miner.Walker = *walker.NewWalker(conf, musk.MakeMaxUniformWalk(Next, miner))
	return miner
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	errors.Logf("INFO", "Customize creation")
	return (w.Walker).Mine(dt, rptr)
}

func Next(ctx interface{}, cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.Children()
	if err != nil {
		return nil, err
	}
	parents, err := cur.Parents()
	if err != nil {
		return nil, err
	}
	adjs := append(kids, parents...)
	// teleports, err := 
	errors.Logf("DEBUG", "cur %v parents %v kids %v adjs %v", cur, len(parents), len(kids), len(adjs))
	prs, err := transPrs(cur, adjs)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return adjs[i], nil
}

func transPrs(u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := weight(u, v)
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

func weight(u, v lattice.Node) (float64, error) {
	umax, err := u.Maximal()
	if err != nil {
		return 0, err
	}
	vmax, err := v.Maximal()
	if err != nil {
		return 0, err
	}
	udeg, err := u.AdjacentCount()
	if err != nil {
		return 0, err
	}
	vdeg, err := v.AdjacentCount()
	if err != nil {
		return 0, err
	}
	if umax && vmax {
		return 0, nil
	} else if !umax && vmax {
		return 1.0 / float64(vdeg), nil
	} else if umax && !vmax {
		return 1.0 / float64(udeg), nil
	} else {
		return 1.0, nil
	}
}

