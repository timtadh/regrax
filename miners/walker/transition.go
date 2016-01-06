package walker

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
	prs := make([]float64, 0, len(adjs))
	for _, wght := range weights {
		prs = append(prs, wght/total)
	}
	return prs, nil
}

func Transition(cur lattice.Node, adjs []lattice.Node, weight Weight) (lattice.Node, error) {
	if len(adjs) <= 0 {
		return nil, nil
	}
	if len(adjs) == 1 {
		return adjs[0], nil
	}
	prs, err := TransitionPrs(cur, adjs, weight)
	if err != nil {
		return nil, err
	}
	i := stats.WeightedSample(prs)
	return adjs[i], nil
}

