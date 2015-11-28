package ospace

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/stats"
)


func UniformWalk(w *walker.Walker) (chan lattice.Node, chan error) {
	nodes := make(chan lattice.Node)
	errs := make(chan error)
	count := 0
	go func() {
		cur := w.Start[rand.Intn(len(w.Start))]
		for count < w.Config.Samples {
			count++
			nodes <- cur
			next, err := Next(w, cur)
			if err != nil {
				errs <- err
				break
			}
			if next == nil {
				errs <- errors.Errorf("next was nil!!")
				break
			}
			cur = next
		}
		close(nodes)
		close(errs)
	}()
	return nodes, errs
}

func Next(w *walker.Walker, cur lattice.Node) (lattice.Node, error) {
	kids, err := cur.Children(w.Config.Support, w.Dt)
	if err != nil {
		return nil, err
	}
	parents, err := cur.Parents(w.Config.Support, w.Dt)
	if err != nil {
		return nil, err
	}
	adjs := append(kids, parents...)
	errors.Logf("DEBUG", "cur %v parents %v kids %v adjs %v", cur, len(parents), len(kids), len(adjs))
	prs, err := transPrs(w, cur, adjs)
	if err != nil {
		return nil, err
	}
	adjs = append(adjs, cur)
	prs = append(prs, selfPr(prs))
	i := stats.WeightedSample(prs)
	return adjs[i], nil
}

func selfPr(prs []float64) float64 {
	return 1.0 - stats.Sum(prs)
}

func transPrs(w *walker.Walker, u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	prs := make([]float64, 0, len(adjs))
	for _, v := range adjs {
		wght, err := weight(w, u, v)
		if err != nil {
			return nil, err
		}
		// errors.Logf("DEBUG", "u %v, v %v, weight: %v", u, v, w)
		prs = append(prs, 1.0/wght)
	}
	return prs, nil
}

func weight(w *walker.Walker, u, v lattice.Node) (float64, error) {
	udeg, err := u.AdjacentCount(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	vdeg, err := v.AdjacentCount(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	return max(float64(udeg), float64(vdeg)), nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

