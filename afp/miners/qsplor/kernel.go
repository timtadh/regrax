package qsplor

import (
	"github.com/timtadh/sfp/stats"
)

type Kernel [][]float64

func (k Kernel) Mean(i int) float64 {
	mean, _ := stats.Mean(stats.Srange(len(k)), func(j int) float64 {
		return k[i][j]
	})
	return mean
}

func kernel(items []int, f func(i, j int) float64) Kernel {
	scores := make(Kernel, len(items))
	for i := range scores {
		scores[i] = make([]float64, len(items))
	}
	for x, i := range items {
		for y, j := range items {
			if i == j {
				scores[x][y] = 0
			} else {
				scores[x][y] = f(i, j)
			}
		}
	}
	return scores
}
