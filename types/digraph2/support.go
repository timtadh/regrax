package digraph2

import (
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)

func (n *Node) support(V int, embs []*subgraph.Embedding) int {
	return n.mni(V, embs)
}

func (n *Node) mni(V int, embs []*subgraph.Embedding) int {
	sets := make([]map[int]bool, V)
	for _, emb := range embs {
		for e := emb; e != nil; e = e.Prev {
			set := sets[e.SgIdx]
			if set == nil {
				set = make(map[int]bool)
				sets[e.SgIdx] = set
			}
			set[e.EmbIdx] = true
		}
	}
	min := -1
	for _, set := range sets {
		if min < 0 || len(set) < min {
			min = len(set)
		}
	}
	return min
}
