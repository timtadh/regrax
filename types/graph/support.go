package graph

import (
)

import (
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
)


func VertexSets(sgs SubGraphs) []*set.MapSet {
	if len(sgs) == 0 {
		return make([]*set.MapSet, 0)
	}
	sets := make([]*set.MapSet, 0, len(sgs[0].V))
	for i := range sgs[0].V {
		set := set.NewMapSet(set.NewSortedSet(len(sgs)))
		for j, sg := range sgs {
			id := types.Int(sg.V[i].Id)
			if !set.Has(id) {
				set.Put(id, j)
			}
		}
		sets = append(sets, set)
	}
	return sets
}

func MinImgSupported(sgs SubGraphs) SubGraphs {
	if len(sgs) <= 1 {
		return sgs
	}
	sets := VertexSets(sgs)
	arg, size := lattice.Min(lattice.Srange(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	supported := make(SubGraphs, 0, int(size)+1)
	for sgIdx, next := sets[arg].Values()(); next != nil; sgIdx, next = next() {
		idx := sgIdx.(int)
		supported = append(supported, sgs[idx])
	}
	return supported
}

