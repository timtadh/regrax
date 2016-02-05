package uniprox

import (
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
	"github.com/timtadh/sfp/stores/bytes_float"
)

type Walker struct {
	walker.Walker
	EstimatingWalks int
	Ests bytes_float.MultiMap
}

func NewWalker(conf *config.Config, estimatingWalks int) (*Walker, error) {
	ests, err := conf.BytesFloatMultiMap("uniprox-weight-ests")
	if err != nil {
		return nil, err
	}
	miner := &Walker{
		EstimatingWalks: estimatingWalks,
		Ests: ests,
	}
	miner.Walker = *walker.NewWalker(conf, absorbing.MakeAbsorbingWalk(absorbing.MakeSample(miner), make(chan error)))
	return miner, nil
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	return (w.Walker).Mine(dt, rptr)
}

func (w *Walker) Next(cur lattice.Node) (lattice.Node, error) {
	ismax, err := cur.Maximal(); 
	if err != nil {
		return nil, err
	} else if ismax {
		return nil, nil
	}
	kids, err := cur.CanonKids()
	if err != nil {
		return nil, err
	} else if len(kids) == 0 {
		errors.Logf("DEBUG", "cur %v is not max but no canon kids", cur)
		errors.Logf("DEBUG", "restarting walk from bottom lattice node")
		return w.Dt.Empty(), nil
	}
	errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	next, err := walker.Transition(cur, kids, w.weight)
	// panic("stop")
	return next, err
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
		est = depth * max(diameter, 1)
		if est >= 0 {
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
		return 0, 0, nil
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
	diameter = float64(maxTail.Level() - anc.Level())
	depth = float64(maxDepth)
	return depth, diameter, nil
}

func (w *Walker) walkFrom(v lattice.Node) (path []lattice.Node, err error) {
	weight := func(_, a lattice.Node) (float64, error) {
		var est float64
		if ismax, err := a.Maximal(); err != nil {
			return 0, err
		} else if !ismax {
			/*
			kids, err := a.CanonKids()
			if err != nil {
				return 0, err
			}
			est = float64(len(kids))*/
			est = 1.0
		} else if v.Pattern().Level() >= w.Dt.MinimumLevel() {
			est = 1.0
		} else {
			est = 0.01
		}
		return est, nil
	}
	transition := func(c lattice.Node) (lattice.Node, bool, error) {
		kids, err := c.CanonKids()
		if err != nil {
			return nil, false, err
		}
		if len(kids) == 0 {
			if ismax, err := c.Maximal(); err != nil {
				return nil, false, err
			} else if !ismax {
				// restart walk
				return nil, true, nil
			}
		}
		next, err := walker.Transition(c, kids, weight)
		return next, false, err
	}
	c := v
	n, _, err := transition(c)
	if err != nil {
		return nil, err
	}
	path = append(path, c)
	restarts := 0
	for n != nil {
		var restart bool
		c = n
		path = append(path, c)
		n, restart, err = transition(c)
		if err != nil {
			return nil, err
		} else if restart && restarts > 5 {
			return []lattice.Node{v}, nil
		} else if restart {
			restarts++
			errors.Logf("DEBUG", "restart(%v) to beginning node %v", restarts, v)
			errors.Logf("DEBUG", "restart(%v) from %v", restarts, c)
			kids, err := c.Children()
			if err != nil {
				return nil, err
			}
			for _, kid := range kids {
				errors.Logf("DEBUG", "restart(%v) kid %v", restarts, kid)
			}
			n = v
			path = make([]lattice.Node, 0, 10)
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
