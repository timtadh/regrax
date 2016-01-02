package uniprox

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/absorbing"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/stats"
	"github.com/timtadh/sfp/stores/bytes_float"
)

type Walker struct {
	walker.Walker
	Ests bytes_float.MultiMap
}

func NewWalker(conf *config.Config) (*Walker, error) {
	ests, err := conf.BytesFloatMultiMap("uniprox-weight-ests")
	if err != nil {
		return nil, err
	}
	miner := &Walker{
		Ests: ests,
	}
	miner.Walker = *walker.NewWalker(conf, absorbing.MakeAbsorbingWalk(absorbing.MakeSample(miner), make(chan error)))
	return miner, nil
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	return (w.Walker).Mine(dt, rptr)
}

func (w *Walker) Next(cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.CanonKids()
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	if len(kids) <= 0 {
		return nil, nil
	}
	if len(kids) == 1 {
		return kids[0], nil
	}
	prs, err := w.transPrs(cur, kids)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return kids[i], nil
}

func (w *Walker) transPrs(u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	errors.Logf("DEBUG", "u %v", u)
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := w.weight(v)
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

func (w *Walker) weight(v lattice.Node) (float64, error) {
	label := v.Pattern().Label()
	if has, err := w.Ests.Has(label); err != nil {
		return 0, err
	} else if has {
		var est float64
		err := w.Ests.DoFind(label, func(_ []byte, f float64) error {
			est = f
			return nil
		})
		if err != nil {
			return 0, err
		}
		return est, nil
	}
	depth, diameter, err := w.estimateDepthDiameter(v, 10)
	if err != nil {
		return 0, err
	}
	var est float64 = depth * max(diameter, .01)
	err = w.Ests.Add(label, est)
	if err != nil {
		return 0, err
	}
	errors.Logf("INFO", "node %v depth %v diameter %v est %v", v, depth, diameter, est)
	return est, nil
}

func (w *Walker) estimateDepthDiameter(v lattice.Node, walks int) (depth, diameter float64, err error) {
	var maxDepth int = 0
	var maxTail lattice.Pattern = nil
	tails := set.NewSortedSet(10)
	for i := 0; i < walks; i++ {
		var path []lattice.Node = nil
		var err error = nil
		if i < walks / 2 {
			path, err = w.walkFrom(v)
		} else {
			path, err = w.quickWalkFrom(v)
		}
		if err != nil {
			return 0, 0, err
		}
		tail := path[len(path)-1].Pattern()
		tails.Add(tail)
		if len(path) > maxDepth {
			maxDepth = len(path)
			maxTail = tail
		}
	}
	errors.Logf("DEBUG", "v %v got depth %v", v, maxDepth)
	diameter = 0.0
	// anc := maxTail
	for i := 0; i < 3; i++ {
		t, err := tails.Random()
		if err != nil {
			return 0, 0, err
		}
		tail := t.(lattice.Pattern)
		a := maxTail.Distance(tail)
		if a > diameter {
			diameter = a
		}
		/*
		a := maxTail.CommonAncestor(tail)
		if a.Level() < anc.Level() {
			anc = a
		}*/
	}
	// diameter = float64(maxTail.Level() - anc.Level())
	depth = float64(maxDepth)
	return depth, diameter, nil
}

func (w *Walker) quickWalkFrom(v lattice.Node) (path []lattice.Node, err error) {
	transition := func(c lattice.Node) (lattice.Node, error) {
		kids, err := c.CanonKids()
		if err != nil {
			return nil, err
		}
		if len(kids) <= 0 {
			return nil, nil
		}
		return kids[rand.Intn(len(kids))], nil
	}
	c := v
	n, err := transition(c)
	if err != nil {
		return nil, err
	}
	path = append(path, c)
	for n != nil {
		c = n
		n, err = transition(c)
		if err != nil {
			return nil, err
		}
		path = append(path, c)
	}
	return path, nil
}

func (w *Walker) walkFrom(v lattice.Node) (path []lattice.Node, err error) {
	weight := func(a lattice.Node) (float64, error) {
		odeg, err := a.ChildCount()
		if err != nil {
			return 0, err
		}
		return float64(odeg) + 1, nil
	}
	prs := func(u lattice.Node, adjs []lattice.Node) ([]float64, error) {
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
	transition := func(c lattice.Node) (lattice.Node, error) {
		kids, err := c.CanonKids()
		if err != nil {
			return nil, err
		}
		if len(kids) <= 0 {
			return nil, nil
		}
		if len(kids) == 1 {
			return kids[0], nil
		}
		prs, err := prs(c, kids)
		if err != nil {
			return nil, err
		}
		i := stats.WeightedSample(prs)
		return kids[i], nil
	}
	c := v
	n, err := transition(c)
	if err != nil {
		return nil, err
	}
	path = append(path, c)
	for n != nil {
		c = n
		n, err = transition(c)
		if err != nil {
			return nil, err
		}
		path = append(path, c)
	}
	return path, nil
}

func max(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}
