package qsplor

import (
	"sync"
)

import (
	"github.com/timtadh/regrax/stats"
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
	var wg sync.WaitGroup
	for x, i := range items {
		wg.Add(1)
		go func(x, i int) {
			defer wg.Done()
			for y, j := range items {
				if i == j {
					scores[x][y] = 0
				} else {
					scores[x][y] = f(i, j)
				}
			}
		}(x, i)
	}
	wg.Wait()
	return scores
}
