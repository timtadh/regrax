package graph

import (
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/stats"
)

func vertexMapSets(sgs SubGraphs) []*set.MapSet {
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

func subgraphVertexSets(sgs SubGraphs) []*set.SortedSet {
	if len(sgs) == 0 {
		return make([]*set.SortedSet, 0)
	}
	sets := make([]*set.SortedSet, 0, len(sgs))
	for _, sg := range sgs {
		s := set.NewSortedSet(len(sgs[0].V))
		for i := range sg.V {
			s.Add(types.Int(sg.V[i].Id))
		}
		sets = append(sets, s)
	}
	return sets
}

func MinImgSupported(sgs SubGraphs) SubGraphs {
	if len(sgs) <= 1 {
		return sgs
	}
	sets := vertexMapSets(sgs)
	arg, size := stats.Min(stats.Srange(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	supported := make(SubGraphs, 0, int(size)+1)
	for sgIdx, next := sets[arg].Values()(); next != nil; sgIdx, next = next() {
		idx := sgIdx.(int)
		supported = append(supported, sgs[idx])
	}
	return supported
}

func DedupSupported(sgs SubGraphs) SubGraphs {
	labels := set.NewSortedSet(len(sgs))
	graphs := make(SubGraphs, 0, len(sgs))
	for _, sg := range sgs {
		label := types.ByteSlice(sg.Serialize())
		if !labels.Has(label) {
			labels.Add(label)
			graphs = append(graphs, sg)
		}
	}
	return graphs
}

func MaxIndepSupported(sgs SubGraphs) SubGraphs {
	if len(sgs) <= 1 {
		return sgs
	}
	sets := subgraphVertexSets(sgs)
	perms := stats.Permutations(len(sets))
	arg, max := stats.Max(stats.Srange(len(perms)), func(idx int) float64 {
		perm := make([]*set.SortedSet, 0, len(sets))
		for _, idx := range perms[idx] {
			perm = append(perm, sets[idx])
		}
		return float64(intersect(perm).Size())
	})
	vids := set.NewSortedSet(int(max))
	nonOverlapping := make(SubGraphs, 0, len(sgs))
	for _, idx := range perms[arg] {
		s := sets[idx]
		if !vids.Overlap(s) {
			errors.Logf("DEBUG", "sgs %v %v %v", len(sgs), idx, len(sets))
			nonOverlapping = append(nonOverlapping, sgs[idx])
			for v, next := s.Items()(); next != nil; v, next = next() {
				item := v.(types.Int)
				if err := vids.Add(item); err != nil {
					panic(err)
				}
			}
		}
	}
	return nonOverlapping
}

func intersect(sets []*set.SortedSet) *set.SortedSet {
	s := sets[0]
	for i := 1; i < len(sets); i++ {
		x, _ := s.Intersect(sets[i])
		s = x.(*set.SortedSet)
	}
	return s
}


