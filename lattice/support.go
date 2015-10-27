package lattice

import (
)

import (
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/data-structures/set"
)


type RawSupport struct{}

func (s RawSupport) Supported(embeddings []Embedding) []Embedding {
	return embeddings
}


type MinImageSupport struct{}

func (s MinImageSupport) Supported(embeddings []Embedding) []Embedding {
	if len(embeddings) <= 1 {
		return embeddings
	}
	sets := componentSets(embeddings)
	arg, size := min(srange(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	supported := make([]Embedding, 0, int(size)+1)
	for embedsIdx, next := sets[arg].Values()(); next != nil; embedsIdx, next = next() {
		idx := embedsIdx.(int)
		supported = append(supported, embeddings[idx])
	}
	return supported
}


func componentSets(embeds []Embedding) []*set.MapSet {
	if len(embeds) == 0 {
		return make([]*set.MapSet, 0)
	}
	C := len(embeds[0].Components())
	sets := make([]*set.MapSet, 0, C)
	for i := 0; i < C; i++ {
		set := set.NewMapSet(set.NewSortedSet(len(embeds)))
		for j, e := range embeds {
			id := types.Int(e.Components()[i])
			if !set.Has(id) {
				set.Put(id, j)
			}
		}
		sets = append(sets, set)
	}
	return sets
}

