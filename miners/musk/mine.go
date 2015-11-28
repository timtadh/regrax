package musk

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


func MaxUniformWalk(w *walker.Walker) (chan lattice.Node, chan bool, chan error) {
	samples := make(chan lattice.Node)
	terminate := make(chan bool)
	errs := make(chan error)
	go func() {
		cur := w.Start[rand.Intn(len(w.Start))]
		loop: for {
			var sampled lattice.Node = nil
			for sampled == nil {
				if ismax, err := cur.Maximal(w.Config.Support, w.Dt); err != nil {
					errs <- err
					break loop
				} else if ismax {
					sampled = cur
				}
				next, err := Next(w, cur)
				if err != nil {
					errs <- err
					break loop
				}
				if next == nil {
					errs <- errors.Errorf("next was nil!!")
					break loop
				}
				cur = next
			}
			select {
			case <-terminate:
				break loop
			case samples<-sampled:
			}
		}
		close(samples)
		close(errs)
		close(terminate)
	}()
	return samples, terminate, errs
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
	i := stats.WeightedSample(prs)
	return adjs[i], nil
}

func transPrs(w *walker.Walker, u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := weight(w, u, v)
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

func weight(w *walker.Walker, u, v lattice.Node) (float64, error) {
	umax, err := u.Maximal(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	vmax, err := v.Maximal(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	udeg, err := u.AdjacentCount(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	vdeg, err := v.AdjacentCount(w.Config.Support, w.Dt)
	if err != nil {
		return 0, err
	}
	if umax && vmax {
		return 0, nil
	} else if !umax && vmax {
		return 1.0/float64(vdeg), nil
	} else if umax && !vmax {
		return 1.0/float64(udeg), nil
	} else {
		return 1.0, nil
	}
}

