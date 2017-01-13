package subgraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/exc"
)

import (
	"github.com/timtadh/sfp/types/digraph/digraph"
)

func (sg *SubGraph) EstimateMatch(indices *digraph.Indices) (match float64, csg *SubGraph, err error) {
	csg = sg
	for len(csg.E) >= 0 {
		found, chain, maxEid, _ := csg.Embedded(indices)
		if false {
			errors.Logf("INFO", "found: %v %v %v %v", found, chain, maxEid, nil)
		}
		if found {
			if len(sg.E) == 0 {
				match = 1
			} else {
				match = float64(maxEid)/float64(len(sg.E))
			}
			return match, csg, nil
		}
		var b *Builder
		connected := false
		eid := maxEid
		if len(csg.E) == 1 {
			break
		}
		for !connected && eid >= 0 && eid < len(chain) {
			b = csg.Builder()
			if err := b.RemoveEdge(chain[eid]); err != nil {
				return 0, nil, err
			}
			connected = b.Connected()
			eid++
		}
		if !connected {
			break
		}
		csg = b.Build()
	}
	return 0, EmptySubGraph(), nil
}

func (sg *SubGraph) Embedded(indices *digraph.Indices) (found bool, edgeChain []int, largestEid int, longest *IdNode) {
	type entry struct {
		ids *IdNode
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	largestEid = -1
	for startIdx := 0; startIdx < len(sg.V); startIdx++ {
		// startIdx := sg.searchStartingPoint(MostExtensions, indices, nil)
		chain := sg.edgeChain(indices, nil, startIdx)
		vembs := sg.startEmbeddings(indices, startIdx)
		stack := make([]entry, 0, len(vembs)*2)
		for _, vemb := range vembs {
			stack = append(stack, entry{vemb, 0})
		}
		if false {
			errors.Logf("DEBUG", "stack %v", stack)
		}
		for len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
			if i.eid > largestEid {
				edgeChain = chain
				longest = i.ids
				largestEid = i.eid
			}
			if i.eid >= len(chain) {
				return true, chain, i.eid, i.ids
			} else {
				sg.extendEmbedding(indices, i.ids, &sg.E[chain[i.eid]], nil, func(ext *IdNode) {
					stack = append(stack, entry{ext, i.eid + 1})
				})
			}
		}
	}
	return false, edgeChain, largestEid, longest
}

func (sg *SubGraph) VisualizeEmbedding(indices *digraph.Indices, labels *digraph.Labels) (dotty string, err error) {
	err = exc.Try(func(){
		g := FromGraph(indices.G).Build()
		edge := func(eid int, vids map[int]int) int {
			e := &sg.E[eid]
			for eid := range g.E {
				if vids[e.Src] == g.E[eid].Src && vids[e.Targ] == g.E[eid].Targ && e.Color == g.E[eid].Color {
					return eid
				}
			}
			panic(exc.Errorf("unreachable").Exception())
		}
		found, edgeChain, eid, ids := sg.Embedded(indices)
		if !found && len(sg.V) > 0 {
			exc.Throwf("%v not found in graph!", sg)
		}
		vids := make(map[int]int)
		vidSet := make(map[int]bool)
		eidSet := make(map[int]bool)
		hv := make(map[int]bool)
		he := make(map[int]bool)
		for c := ids; c != nil; c = c.Prev {
			vids[c.Idx] = c.Id
			vidSet[c.Id] = true
		}
		for vidx := range g.V {
			if !vidSet[vidx] {
				hv[vidx] = true
			}
		}
		for ceid, eidx := range edgeChain {
			if ceid >= eid {
				continue
			}
			eidSet[edge(eidx, vids)] = true
		}
		for eidx := range g.E {
			if !eidSet[eidx] {
				he[eidx] = true
			}
		}
		dotty = g.Dotty(labels, hv, he)
	}).Error()
	return dotty, err
}
