package walker

import (
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stats"
)


type Weight func(u, v lattice.Node) (float64, error)

func TransitionPrs(u lattice.Node, adjs []lattice.Node, weight Weight) ([]float64, error) {
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
	if total == 0 {
		return nil, nil
	}
	prs := make([]float64, 0, len(adjs))
	for _, wght := range weights {
		prs = append(prs, wght/total)
	}
	return prs, nil
}

func Transition(cur lattice.Node, adjs []lattice.Node, weight Weight) (float64, lattice.Node, error) {
	if len(adjs) <= 0 {
		return 1, nil, nil
	}
	if len(adjs) == 1 {
		return 1, adjs[0], nil
	}
	prs, err := TransitionPrs(cur, adjs, weight)
	if err != nil {
		return 0, nil, err
	} else if prs == nil {
		return 1, nil, nil
	}
	s := stats.Round(stats.Sum(prs), 3)
	if s != 1.0 {
		weights := make([]float64, 0, len(adjs))
		for _, v := range adjs {
			wght, _ := weight(cur, v)
			weights = append(weights, wght)
		}
		return 0, nil, errors.Errorf("sum(%v) (%v) != 1.0 from %v", prs, s, weights)
	}
	i := stats.WeightedSample(prs)
	return prs[i], adjs[i], nil
}

