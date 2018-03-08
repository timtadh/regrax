package qsplor

import (
	"math/rand"
)

import (
	"github.com/timtadh/regrax/stats"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/types/digraph"
)

type Scorer interface {
	Score(lattice.Node, []lattice.Node) float64
	Kernel([]lattice.Node, []int) Kernel
}

var Scorers = map[string]Scorer {
	"random": &RandomScore{},
	"walk-kernel": &WalkKernel{},
}


type RandomScore struct{}
func (r *RandomScore) Score(n lattice.Node, population []lattice.Node) float64 { return rand.Float64() }
func (r *RandomScore) Kernel(population []lattice.Node, s []int) Kernel { return nil }


type WalkKernel struct{}

func (q *WalkKernel) Score(n lattice.Node, population []lattice.Node) float64 {
	sampleSize := 10
	mean, _ := stats.Mean(stats.Sample(sampleSize, len(population)), func(i int) float64 {
		o := population[i]
		return n.(*digraph.EmbListNode).Pat.Metric(o.(*digraph.EmbListNode).Pat)
	})
	return mean
}

func (q *WalkKernel) Kernel(population []lattice.Node, s []int) Kernel {
	return kernel(s, func(i, j int) float64 {
		return population[i].(*digraph.EmbListNode).Pat.Metric(population[j].(*digraph.EmbListNode).Pat)
	})
}

