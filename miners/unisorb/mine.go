package unisorb

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
				if ismax, err := cur.Maximal(); err != nil {
					errs <- err
					break loop
				} else if ismax {
					sampled = cur
					cur = w.Start[rand.Intn(len(w.Start))]
				} else {
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
	kids, err := cur.Children()
	if err != nil {
		return nil, err
	}
	// errors.Logf("DEBUG", "cur %v kids %v", cur, len(kids))
	prs, err := transPrs(w, cur, kids)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return kids[i], nil
}

func transPrs(w *walker.Walker, u lattice.Node, adjs []lattice.Node) ([]float64, error) {
	weights := make([]float64, 0, len(adjs))
	var total float64 = 0
	for _, v := range adjs {
		wght, err := weight(w, v)
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

func weight(w *walker.Walker, v lattice.Node) (float64, error) {
	vmax, err := v.Maximal()
	if err != nil {
		return 0, err
	}
	vdeg, err := v.ParentCount()
	if err != nil {
		return 0, err
	}
	if vmax {
		return 1.0/float64(vdeg), nil
	} else {
		return 1.0, nil
	}
}

