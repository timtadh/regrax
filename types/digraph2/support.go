package digraph2

import (
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)

func (n *Node) support(embs []*subgraph.Embedding) int {
	return n.mni(embs)
}

func (n *Node) mni(embs []*subgraph.Embedding) int {
	sets := make(map[int]map[int]bool)
	for _, emb := range embs {
		for e := emb; e != nil; e = e.Prev {
			set, has := sets[e.SgIdx]
			if !has {
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
