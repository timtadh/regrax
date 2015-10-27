package itemset

import (
	"io"
)

import (
	"github.com/timtadh/sfp/lattice"
)


type MakeLoader func(*ItemSets) lattice.Loader


type ItemSets struct {
	InvertedIndex map[int][]int
	FrequentItems []int
	makeLoader MakeLoader
}

func NewItemSets(makeLoader MakeLoader) *ItemSets {
	return &ItemSets{
		InvertedIndex: make(map[int][]int),
		FrequentItems: make([]int, 0, 10),
		makeLoader: makeLoader,
	}
}

func (i *ItemSets) Metric() lattice.SupportMetric {
	return lattice.RawSupport{}
}

func (i *ItemSets) Loader() lattice.Loader {
	return i.makeLoader(i)
}


type IntLoader struct {
	sets *ItemSets
}

func NewIntLoader(sets *ItemSets) lattice.Loader {
	return &IntLoader{
		sets: sets,
	}
}

func (l *IntLoader) StartingPoints(input io.Reader, support int) []lattice.Node {
	panic("unimplemented")
}

