package subgraph

import (
	"github.com/timtadh/data-structures/errors"
)

func (sg *SubGraph) EstimateMatch(indices *Indices) (match float64, csg *SubGraph, err error) {
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
		for !connected && eid < len(chain) {
			b = csg.Builder()
			if err := b.RemoveEdge(chain[eid]); err != nil {
				return 0, nil, err
			}
			connected = b.Connected()
			eid++
		}
		csg = b.Build()
	}
	return 0, csg, errors.Errorf("unreachable")
}

func (sg *SubGraph) Embedded(indices *Indices) (found bool, edgeChain []int, largestEid int, longest *IdNode) {
	type entry struct {
		ids *IdNode
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	for startIdx := 0; startIdx < len(sg.V); startIdx++ {
		startIdx := sg.mostExts(indices, nil)
		chain := sg.edgeChain(indices, nil, startIdx)
		vembs := sg.startEmbeddings(indices, startIdx)
		stack := make([]entry, 0, len(vembs)*2)
		for _, vemb := range vembs {
			stack = append(stack, entry{vemb, 0})
		}
		if false {
			errors.Logf("DEBUG", "stack %v", stack)
		}
		largestEid = -1
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
