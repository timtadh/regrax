package digraph

import (
	"encoding/json"
)

import (
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/list"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stats"
)

type Supported func(dt *Digraph, sgs SubGraphs) (SubGraphs, error)

func MakeTxSupported(attrName string) Supported {
	return func(dt *Digraph, sgs SubGraphs) (SubGraphs, error) {
		sgs, err := MinImgSupported(dt, sgs)
		if err != nil {
			return nil, err
		}
		supported := make(SubGraphs, 0, len(sgs))
		seen := set.NewSortedSet(len(sgs))
		for _, sg := range sgs {
			if len(sg.V) <= 0 {
				supported = append(supported, sg)
				continue
			}
			vid := int32(sg.V[0].Id)
			err := dt.NodeAttrs.DoFind(vid, func(_ int32, attrs map[string]interface{}) error {
				attr, has := attrs[attrName];
				if !has {
					return errors.Errorf("subgraph %v vertex %v did not have attr %v in attrs %v", sg, sg.V[0], attrName, attrs)
				}
				switch a := attr.(type) {
				case json.Number:
					if n, err := a.Int64(); err != nil {
						return errors.Errorf("type float is not suppported as an attr value for tx id %v %v for %v", attrName, attr, sg)
					} else {
						if !seen.Has(types.Int64(n)) {
							seen.Add(types.Int64(n))
							supported = append(supported, sg)
						}
					}
				case map[string]interface{}:
					return errors.Errorf("type map[string]interface{} is not suppported as an attr value for tx id")
				case []interface{}:
					return errors.Errorf("type []interface{} is not suppported as an attr value for tx id")
				case bool:
					return errors.Errorf("type bool is not suppported as an attr value for tx id")
				case nil:
					return errors.Errorf("type nil is not suppported as an attr value for tx id")
				case string:
					if !seen.Has(types.String(a)) {
						seen.Add(types.String(a))
						supported = append(supported, sg)
					}
				default:
					return errors.Errorf("unexpected type %T for %v in %v", attr, attr, sg)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return supported, nil
	}
}

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

func MinImgSupported(dt *Digraph, sgs SubGraphs) (SubGraphs, error) {
	if len(sgs) <= 1 {
		return sgs, nil
	}
	sets := vertexMapSets(sgs)
	// errors.Logf("MIN-IMAGE", "sets: %v", sets)
	arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	if int(size) == len(sgs) {
		return sgs, nil
	}
	supported := make(SubGraphs, 0, int(size)+1)
	for sgIdx, next := sets[arg].Values()(); next != nil; sgIdx, next = next() {
		idx := sgIdx.(int)
		supported = append(supported, sgs[idx])
	}
	return supported, nil
}

func Dedup(sgs SubGraphs) SubGraphs {
	vertices := func(sg *goiso.SubGraph) *list.List {
		l := list.New(len(sg.V))
		for i := range sg.V {
			l.Append(types.Int(sg.V[i].Id))
		}
		return l
	}
	labels := hashtable.NewLinearHash()
	graphs := make(SubGraphs, 0, len(sgs))
	for _, sg := range sgs {
		// label := types.ByteSlice(sg.Serialize())
		label := vertices(sg)
		if !labels.Has(label) {
			labels.Put(label, nil)
			graphs = append(graphs, sg)
		}
	}
	return graphs
}

func MaxIndepSupported(dt *Digraph, sgs SubGraphs) (SubGraphs, error) {
	sgs, err := MinImgSupported(dt, sgs)
	if err != nil {
		return nil, err
	}
	if len(sgs) <= 1 {
		return sgs, nil
	}
	sets := subgraphVertexSets(sgs)
	// errors.Logf("MAX-INDEP", "sets: %v", sets)
	max := -1
	var maxPerm []int = nil
	var maxNonOverlap []int = nil
	stats.Permutations(len(sets), func(perm []int) (dobreak bool) {
		nonOverlap := nonOverlapping(perm, sets)
		size := len(nonOverlap)
		if size > max || maxPerm == nil {
			max = size
			maxPerm = perm
			maxNonOverlap = nonOverlap
		}
		if size == len(sgs) {
			return true
		}
		// errors.Logf("MAX-INDEP", "perm %v %v %v :: max: %v %v %v", perm, size, nonOverlap, maxPerm, max, maxNonOverlap)
		return false
	})
	nonOverlapping := make(SubGraphs, 0, len(maxNonOverlap))
	for _, idx := range maxNonOverlap {
		nonOverlapping = append(nonOverlapping, sgs[idx])
	}
	return nonOverlapping, nil
}

func nonOverlapping(perm []int, sets []*set.SortedSet) []int {
	nonOverlapping := make([]int, 0, len(sets))
	idxs := set.NewSortedSet(len(sets))
	for _, idx := range perm {
		s := sets[idx]
		if !idxs.Overlap(s) {
			nonOverlapping = append(nonOverlapping, idx)
			for v, next := s.Items()(); next != nil; v, next = next() {
				item := v.(types.Int)
				if err := idxs.Add(item); err != nil {
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
		x, err := s.Intersect(sets[i])
		if err != nil {
			panic(err)
		}
		s = x.(*set.SortedSet)
	}
	return s
}
