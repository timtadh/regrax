package subgraph

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
