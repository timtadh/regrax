package musk

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners/walker"
)

type Transition func(interface{}, lattice.Node) (lattice.Node, error)

func MakeMaxUniformWalk(next Transition, ctx interface{}) walker.Walk {
	return func(w *walker.Walker) (chan lattice.Node, chan bool, chan error) {
		samples := make(chan lattice.Node)
		terminate := make(chan bool)
		errs := make(chan error)
		go func() {
			cur := w.Dt.Root()
		loop:
			for {
				var sampled lattice.Node = nil
				for sampled == nil {
					if ismax, err := cur.Maximal(); err != nil {
						errs <- err
						break loop
					} else if ismax {
						sampled = cur
					}
					next, err := next(ctx, cur)
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
				samples <- sampled
				if <-terminate {
					break loop
				}
			}
			close(samples)
			close(errs)
		}()
		return samples, terminate, errs
	}
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
	errors.Logf("DEBUG", "cur %v parents %v kids %v adjs %v", cur, len(parents), len(kids), len(adjs))
	_, next, err := walker.Transition(cur, adjs, weight)
	return next, err
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
