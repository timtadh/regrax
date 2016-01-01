package uniprox

import (
	"math/rand"
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
	depth, diameter, err := w.estimateDepthDiameter(v, 25)
	if err != nil {
		return 0, err
	}
	var est float64 = depth * max(diameter, 1)
	err = w.Ests.Add(label, est)
	if err != nil {
		return 0, err
	}
	return est, nil
}

func (w *Walker) estimateDepthDiameter(v lattice.Node, walks int) (depth, diameter float64, err error) {
	var maxDepth int = 0
	var maxTail lattice.Node = nil
	tails := hashtable.NewLinearHash()
	for i := 0; i < walks; i++ {
		path, err := w.walkFrom(v)
		if err != nil {
			return 0, 0, err
		}
		tail := path[len(path)-1]
		tailLabel := types.ByteSlice(tail.Pattern().Label())
		if !tails.Has(tailLabel) {
			tails.Put(tailLabel, tail)
		}
		if len(path) > maxDepth {
			maxDepth = len(path)
			maxTail = tail
		}
	}
	anc := maxTail.Pattern()
	for t, next := tails.Values()(); next != nil; t, next = next() {
		tail := t.(lattice.Node)
		anc = anc.CommonAncestor(tail.Pattern())
	}
	diameter = float64(maxTail.Pattern().Level() - anc.Level())
	depth = float64(maxDepth)
	return depth, diameter, nil
}

func (w *Walker) walkFrom(v lattice.Node) (path []lattice.Node, err error) {
	transition := func(c lattice.Node) (lattice.Node, error) {
		kids, err := c.CanonKids()
		if err != nil {
			return nil, err
		}
		if len(kids) > 0 {
			return kids[rand.Intn(len(kids))], nil
		}
		return nil, nil
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
