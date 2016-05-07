package subgraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/list"
	"github.com/timtadh/data-structures/linked"
	"github.com/timtadh/data-structures/types"
)

import (
	// "github.com/timtadh/sfp/stats"
)

// Tim You Are Here:
// You just ran %s/\*goiso.SubGraph/*Embedding/g
//
// Now it is time to transition this thing over to *Embeddings :check:
// Next it is time to create stores for *Embeddings
// Then it is time to transition types/digraph to *Embeddings

type EmbIterator func() (*Embedding, EmbIterator)

func (sg *SubGraph) Embeddings(indices *Indices) ([]*Embedding, error) {
	embeddings := make([]*Embedding, 0, 10)
	err := sg.DoEmbeddings(indices, func(emb *Embedding) error {
		embeddings = append(embeddings, emb)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return embeddings, nil

}

func (sg *SubGraph) DoEmbeddings(indices *Indices, do func(*Embedding) error) error {
	ei, err := sg.IterEmbeddings(indices, nil)
	if err != nil {
		return err
	}
	for emb, next := ei(); next != nil; emb, next = next() {
		err := do(emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func FilterAutomorphs(it EmbIterator, err error) (ei EmbIterator, _ error) {
	if err != nil {
		return nil, err
	}
	idSet := func(emb *Embedding) *list.Sorted {
		ids := list.NewSorted(len(emb.Ids), true)
		for _, id := range emb.Ids {
			ids.Add(types.Int(id))
		}
		return ids
	}
	seen := hashtable.NewLinearHash()
	ei = func() (emb *Embedding, _ EmbIterator) {
		if it == nil {
			return nil, nil
		}
		for emb, it = it(); it != nil; emb, it = it() {
			ids := idSet(emb)
			// errors.Logf("AUTOMORPH-DEBUG", "emb %v ids %v has %v", emb, ids, seen.Has(ids))
			if !seen.Has(ids) {
				seen.Put(ids, nil)
				return emb, ei
			}
		}
		return nil, nil
	}
	return ei, nil
}

type IdNode struct {
	Id  int
	Idx int
	Prev *IdNode
}

func (ids *IdNode) list(length int) []int {
	l := make([]int, length)
	for c := ids; c != nil; c = c.Prev {
		l[c.Idx] = c.Id
	}
	return l
}

func (sg *SubGraph) IterEmbeddings(indices *Indices, prune func(*IdNode) bool) (ei EmbIterator, err error) {
	type entry struct {
		ids *IdNode
		eid int
	}
	// seen := set.NewSetMap(hashtable.NewLinearHash())
	seen := make(map[int]bool)
	// clean := func(stack []entry, prune func(*FillableEmbeddingBuilder) bool) ([]entry) {
	// 	if prune == nil {
	// 		return stack
	// 	}
	// 	cleaned := make([]entry, 0, len(stack))
	// 	for _, e := range stack {
	// 		if !prune(e.emb) {
	// 			cleaned = append(cleaned, e)
	// 		}
	// 	}
	// 	return cleaned
	// }
	pop := func(stack []entry) (entry, []entry) {
		// remove to enable information maximization stack pop
		return stack[len(stack)-1], stack[0 : len(stack)-1]
		// is super slow because of copy (consider swap delete)
		// sampleSize := 5
		// maxIter := 25
		// unseenCount := func(ids *IdNode) int {
		// 	total := 0
		// 	for c := ids; c != nil; c = c.Prev {
		// 		if _, has := seen[c.Id]; !has {
		// 			total += 1
		// 		}
		// 	}
		// 	return total
		// }
		// var idx int
		// if len(stack) <= maxIter {
		// 	max := -1
		// 	for i, e := range stack {
		// 		c := unseenCount(e.ids)
		// 		if c > max {
		// 			idx = i
		// 			max = c
		// 		}
		// 	}
		// } else {
		// 	idx, _ = stats.Max(append(stats.ReplacingSample(sampleSize + 1, len(stack)-1), len(stack)-1), func(i int) float64 {
		// 		return float64(unseenCount(stack[i].ids))
		// 	})
		// }
		// e := stack[idx]
		// stack[idx] = stack[len(stack) - 1]
		// return e, stack[0 : len(stack) - 1]
	}

	if len(sg.V) == 0 {
		ei = func() (*Embedding, EmbIterator) {
			return nil, nil
		}
		return ei, nil
	}
	startIdx := sg.leastFrequentVertex(indices)
	chain := sg.edgeChain(startIdx)
	vembs := sg.startEmbeddings(indices, startIdx)

	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		stack = append(stack, entry{vemb, 0})
	}

	ei = func() (*Embedding, EmbIterator) {
		for len(stack) > 0 {
			var i entry
			// errors.Logf("DEBUG", "stack %v", len(stack))
			i, stack = pop(stack)
			if prune != nil && prune(i.ids) {
				continue
			}
			// otherwise success we have an embedding we haven't seen
			if i.eid >= len(chain) {
				// check that this is the subgraph we sought
				emb := &Embedding{
					SG: sg,
					Ids: i.ids.list(len(sg.V)),
				}
				// errors.Logf("FOUND", "\n  builder %v %v\n    built %v\n  pattern %v", i.emb.Builder, i.emb.Ids, emb, emb.SG)
				// if !emb.Exists(indices.G) {
				// 	errors.Logf("FOUND", "NOT EXISTS\n  builder %v %v\n    built %v\n  pattern %v", i.emb.Builder, i.emb.Ids, emb, emb.SG)
				// 	panic("wat")
				// }
				for _, id := range emb.Ids {
					seen[id] = true
				}
				if sg.Equals(emb) {
					// sweet we can yield this embedding!
					// stack = clean(stack, prune)
					return emb, ei
				}
				// nope wasn't an embedding drop it
			} else {
				// ok extend the embedding
				// errors.Logf("DEBUG", "\n  extend %v %v %v", i.emb.Builder, i.emb.Ids, chain[i.eid])
				sg.extendEmbedding(indices, i.ids, chain[i.eid], func(ext *IdNode) {
					stack = append(stack, entry{ext, i.eid + 1})
				})
				// errors.Logf("DEBUG", "stack len %v", len(stack))
			}
		}
		return nil, nil
	}
	return ei, nil
}

func (sg *SubGraph) leastFrequentVertex(indices *Indices) int {
	minFreq := -1
	minIdx := -1
	for idx := range sg.V {
		freq := indices.G.ColorFrequency(sg.V[idx].Color)
		if minIdx < 0 || minFreq > freq {
			minFreq = freq
			minIdx = idx
		}
	}
	return minIdx
}

func (sg *SubGraph) startEmbeddings(indices *Indices, startIdx int) []*IdNode {
	color := sg.V[startIdx].Color
	embs := make([]*IdNode, 0, indices.G.ColorFrequency(color))
	for _, gIdx := range indices.ColorIndex[color] {
		embs = append(embs, &IdNode{Id: gIdx, Idx: startIdx})
		// embs = append(embs,
		// 	BuildEmbedding(len(sg.V), len(sg.E)).Fillable().
		// 		Ctx(func(b *FillableEmbeddingBuilder) {
		// 			b.SetVertex(startIdx, color, gIdx)
		// 		}))
	}
	return embs
}

// this is really a breadth first search from the given idx
func (sg *SubGraph) edgeChain(startIdx int) []*Edge {
	edges := make([]*Edge, 0, len(sg.E))
	added := make(map[int]bool, len(sg.E))
	seen := make(map[int]bool, len(sg.V))
	queue := linked.NewUniqueDeque()
	queue.EnqueFront(types.Int(startIdx))
	for queue.Size() > 0 {
		x, err := queue.DequeBack()
		if err != nil {
			errors.Logf("ERROR", "UniqueDeque should never error on Deque\n%v", err)
			panic(err)
		}
		u := int(x.(types.Int))
		if seen[u] {
			continue
		}
		seen[u] = true
		for _, e := range sg.Adj[u] {
			if !added[e] {
				added[e] = true
				edges = append(edges, &sg.E[e])
			}
		}
		for _, e := range sg.Adj[u] {
			queue.EnqueFront(types.Int(sg.E[e].Src))
			queue.EnqueFront(types.Int(sg.E[e].Targ))
		}
	}
	if len(edges) != len(sg.E) {
		panic("assert-fail: len(edges) != len(sg.E)")
	}
	// errors.Logf("DEBUG", "edge chain seen %v", seen)
	// errors.Logf("DEBUG", "edge chain added %v", added)
	// errors.Logf("DEBUG", "edge chain added %v", added)
	return edges
}

func (ids *IdNode) ids(srcIdx, targIdx int) (srcId, targId int) {
	srcId = -1
	targId = -1
	for c := ids; c != nil; c = c.Prev {
		if c.Idx == srcIdx {
			srcId = c.Id
		}
		if c.Idx == targIdx {
			targId = c.Id
		}
	}
	return srcId, targId
}

func (ids *IdNode) String() string {
	items := make([]string, 0, 10)
	for c := ids; c != nil; c = c.Prev {
		items = append(items, fmt.Sprintf("<id: %v, idx: %v>", c.Id, c.Idx))
	}
	ritems := make([]string, len(items))
	idx := len(items)-1
	for _, item := range items {
		ritems[idx] = item
		idx--
	}
	return "{" + strings.Join(ritems, ", ") + "}"
}

func (sg *SubGraph) extendEmbedding(indices *Indices, cur *IdNode, e *Edge, do func(*IdNode)) {
	// errors.Logf("DEBUG", "extend emb %v with %v", cur.Label(), e)
	// exts := ext.NewCollector(-1)
	// exts = make([]*FillableEmbeddingBuilder, 0, 10)

	// src := cur.V[e.Src].Idx
	// targ := cur.V[e.Targ].Idx
	srcId, targId := cur.ids(e.Src, e.Targ)

	if srcId == -1 && targId == -1 {
		panic("src and targ == -1. Which means the edge chain was not connected.")
	} else if srcId != -1 && targId != -1 {
		// both src and targ are in the builder so we can just add this edge
		// errors.Logf("EMB-DEBUG", "    add existing %v", e)
		// if indices.HasEdge(cur.Ids[src], cur.Ids[targ], e.Color) {
		if indices.HasEdge(srcId, targId, e.Color) {
			do(cur)
			// exts = append(exts, cur.Ctx(func(b *FillableEmbeddingBuilder) {
			// 	b.AddEdge(&cur.V[e.Src], &cur.V[e.Targ], e.Color)
			// }))
		}
	} else if srcId != -1 {
		indices.TargsFromSrc(srcId, e.Color, sg.V[e.Targ].Color, cur, func(targId int) {
			// errors.Logf("EMB-DEBUG", "    add targ vertex, %v ke %v", e, ke)
			do(&IdNode{Id: targId, Idx: e.Targ, Prev: cur})
			//exts = append(exts, cur.Copy().Ctx(func(b *FillableEmbeddingBuilder) {
			//	b.SetVertex(e.Targ, sg.V[e.Targ].Color, targ)
			//	b.AddEdge(&b.V[e.Src], &b.V[e.Targ], e.Color)
			//}))
		})
	} else if targId != -1 {
		indices.SrcsToTarg(targId, e.Color, sg.V[e.Src].Color, cur, func(srcId int) {
			do(&IdNode{Id: srcId, Idx: e.Src, Prev: cur})
			// errors.Logf("EMB-DEBUG", "    add src vertex, %v pe %v", e, pe)
			//exts = append(exts, cur.Copy().Ctx(func(b *FillableEmbeddingBuilder) {
			//	b.SetVertex(e.Src, sg.V[e.Src].Color, src)
			//	b.AddEdge(&b.V[e.Src], &b.V[e.Targ], e.Color)
			//}))
		})
	} else {
		panic("unreachable")
	}
}
