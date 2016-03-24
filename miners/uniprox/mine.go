package uniprox

import (
	"bytes"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/graple"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/stores/bytes_float"
)

type Walker struct {
	walker.Walker
	EstimatingWalks int
	Ests            bytes_float.MultiMap
	Prs             bytes_float.MultiMap
	Max             bool
}

func NewWalker(conf *config.Config, estimatingWalks int, max bool) (*Walker, error) {
	ests, err := conf.BytesFloatMultiMap("uniprox-weight-ests")
	if err != nil {
		return nil, err
	}
	prs, err := conf.BytesFloatMultiMap("uniprox-selection-prs")
	if err != nil {
		return nil, err
	}
	miner := &Walker{
		EstimatingWalks: estimatingWalks,
		Ests:            ests,
		Prs:             prs,
		Max:             max,
	}
	miner.Walker = *walker.NewWalker(conf, graple.MakeAbsorbingWalk(graple.MakeSample(miner), make(chan error)))
	return miner, nil
}

func (w *Walker) PrFormatter() lattice.PrFormatter {
	if w.Max {
		return nil
	}
	return NewPrFormatter(w)
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter, fmtr lattice.Formatter) error {
	return (w.Walker).Mine(dt, rptr, fmtr)
}

func (w *Walker) Next(cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.CanonKids()
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	pr, next, err := walker.Transition(cur, kids, w.weight)
	if err != nil {
		return nil, err
	}
	if next == nil && w.Max {
		if ismax, err := cur.Maximal(); err != nil {
			return nil, err
		} else if !ismax {
			return w.Dt.Root(), nil
		}
	}
	return next, w.transitionProbability(cur, next, pr)
}

func (w *Walker) transitionProbability(cur, next lattice.Node, pr float64) error {
	if next == nil {
		return nil
	}
	if has, err := w.Prs.Has(next.Pattern().Label()); err != nil {
		return err
	} else if has {
		return nil
	}
	var curPr float64
	if bytes.Equal(w.Dt.Root().Pattern().Label(), cur.Pattern().Label()) {
		curPr = 1.0
	} else {
		err := w.Prs.DoFind(cur.Pattern().Label(), func(_ []byte, cpr float64) error {
			// errors.Logf("PR", "cur %v curPr %v", cur, cpr)
			curPr = cpr
			return nil
		})
		if err != nil {
			return err
		}
	}
	nextPr := curPr * pr
	// errors.Logf("PR", "adding next %v %v %v", next, next.Pattern().Label(), nextPr)
	err := w.Prs.Add(next.Pattern().Label(), nextPr)
	if err != nil {
		return err
	}
	// errors.Logf("PR", "curPr %v  -> %v ->  nextPr %v", curPr, pr, nextPr)
	return nil
}

func (w *Walker) weight(_, v lattice.Node) (float64, error) {
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
	var est float64
	if ismax, err := v.Maximal(); err != nil {
		return 0, err
	} else if !ismax {
		depth, diameter, err := w.estimateDepthDiameter(v, w.EstimatingWalks)
		if err != nil {
			return 0, err
		}
		est = depth * diameter
		if est >= 1 {
			errors.Logf("DEBUG", "weight %v depth %v diameter %v est %v", v, depth, diameter, est)
		}
	} else if v.Pattern().Level() >= w.Dt.MinimumLevel() {
		est = 1.0
		// errors.Logf("INFO", "node %v is max %v est %v", v, ismax, est)
	} else {
		est = 0.0
		// errors.Logf("INFO", "node %v is max %v but too small v est %v", v, ismax, est)
	}
	err := w.Ests.Add(label, est)
	if err != nil {
		return 0, err
	}
	return est, nil
}

func (w *Walker) estimateDepthDiameter(v lattice.Node, walks int) (depth, diameter float64, err error) {
	if kids, err := v.CanonKids(); err != nil {
		return 0, 0, err
	} else if len(kids) <= 0 {
		return 1, 1, nil
	}
	var maxDepth int = 0
	var maxTail lattice.Pattern = nil
	tails := set.NewSortedSet(10)
	for i := 0; i < walks; i++ {
		var path []lattice.Node = nil
		var err error = nil
		path, err = w.walkFrom(v)
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
	level := maxDepth + v.Pattern().Level()
	if level < w.Dt.MinimumLevel() {
		return 0, 0, nil
	}
	patterns := make([]lattice.Pattern, 0, tails.Size())
	for t, next := tails.Items()(); next != nil; t, next = next() {
		patterns = append(patterns, t.(lattice.Pattern))
	}
	anc, err := CommonAncestor(patterns)
	if err != nil {
		return 0, 0, err
	}
	diameter = float64(maxTail.Level()-anc.Level()) + 1
	depth = float64(maxDepth) + 1
	return depth, diameter, nil
}

func (w *Walker) walkFrom(v lattice.Node) (path []lattice.Node, err error) {
	weight := func(_, a lattice.Node) (float64, error) {
		kids, err := a.CanonKids()
		if err != nil {
			return 0, err
		}
		return float64(len(kids)), nil
	}
	transition := func(c lattice.Node) (lattice.Node, error) {
		kids, err := c.CanonKids()
		if err != nil {
			return nil, err
		}
		_, next, err := walker.Transition(c, kids, weight)
		return next, err
	}
	c := v
	n, err := transition(c)
	if err != nil {
		return nil, err
	}
	path = append(path, c)
	for n != nil {
		c = n
		path = append(path, c)
		n, err = transition(c)
		if err != nil {
			return nil, err
		}
	}
	return path, nil
}

func max(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}
