package premusk

import (
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/musk"
	"github.com/timtadh/sfp/miners/ospace"
	"github.com/timtadh/sfp/miners/reporters"
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

	pConf := w.Config.Copy()
	pConf.Samples = 1000
	premine := walker.NewWalker(pConf, ospace.UniformWalk)
	collector := &reporters.Collector{make([]lattice.Node, 0, 10)}
	pRptr := &reporters.Skip{Skip:10, Reporter:&reporters.Chain{[]miners.Reporter{&reporters.Log{}, collector}}}

	err := premine.Mine(dt, pRptr)
	if err != nil {
		return err
	}

	for i, a := range collector.Nodes {
		for j, b := range collector.Nodes {
			if i == j {
				continue
			}
			key := types.ByteSlice(a.Label())
			var list []lattice.Node
			if w.Teleports.Has(key) {
				val, err := w.Teleports.Get(key)
				if err != nil {
					return err
				}
				list = val.([]lattice.Node)
			} else {
				list = make([]lattice.Node, 0, 10)
			}
			list = append(list, b)
			err = w.Teleports.Put(key, list)
			if err != nil {
				return err
			}
		}
	}
	errors.Logf("DEBUG", "pre-mined %v", collector)

	return (w.Walker).Mine(dt, rptr)
}

func Next(ctx interface{}, cur lattice.Node) (lattice.Node, error) {
	w := ctx.(*Walker)
	kids, err := cur.Children()
	if err != nil {
		return nil, err
	}
	parents, err := cur.Parents()
	if err != nil {
		return nil, err
	}
	adjs := append(kids, parents...)
	teleports, err := w.teleports(cur)
	if err != nil {
		return nil, err
	}
	adjs = append(adjs, teleports...)
	errors.Logf("DEBUG", "cur %v parents %v kids %v teleports %v adjs %v", cur, len(parents), len(kids), len(teleports), len(adjs))
	prs, err := w.transPrs(cur, adjs)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return adjs[i], nil
}

func (w *Walker) teleports(u lattice.Node) ([]lattice.Node, error) {
	key := types.ByteSlice(u.Label())
	if w.Teleports.Has(key) {
		val, err := w.Teleports.Get(key)
		if err != nil {
			return nil, err
		}
		return val.([]lattice.Node), nil
	}
	return []lattice.Node{}, nil
}

func (w *Walker) transPrs(u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := w.weight(u, v)
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

func (w *Walker) weight(u, v lattice.Node) (float64, error) {
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
	utel, err := w.teleports(u)
	if err != nil {
		return 0, err
	}
	udeg += len(utel)
	vdeg, err := v.AdjacentCount()
	if err != nil {
		return 0, err
	}
	vtel, err := w.teleports(v)
	if err != nil {
		return 0, err
	}
	vdeg += len(vtel)
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

