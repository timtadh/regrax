package subgraph

import (
	"fmt"
	// "math/rand"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
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

func (ids *IdNode) idSet(length int) *set.SortedSet {
	s := set.NewSortedSet(length)
	for c := ids; c != nil; c = c.Prev {
		s.Add(types.Int(c.Id))
	}
	return s
}

func (ids *IdNode) has(id, idx int) bool {
	for c := ids; c != nil; c = c.Prev {
		if id == c.Id && idx == c.Idx {
			return true
		}
	}
	return false
}

func (sg *SubGraph) IterEmbeddings(indices *Indices, prune func(*IdNode) bool) (ei EmbIterator, err error) {
	type entry struct {
		ids *IdNode
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}

	if len(sg.V) == 0 {
		ei = func() (*Embedding, EmbIterator) {
			return nil, nil
		}
		return ei, nil
	}
	// startIdx := sg.leastFrequentVertex(indices)
	// startIdx := rand.Intn(len(sg.V))
	// startIdx := sg.mostConnected()
	startIdx := sg.leastConnectedAndExts(indices)
	// if len(sg.E) > 0 {
	// 	errors.Logf("DEBUG", "startIdx %v adj %v freq %v label %v", startIdx, sg.Adj[startIdx], indices.G.ColorFrequency(sg.V[startIdx].Color), indices.G.Colors[sg.V[startIdx].Color])
	// 	leastFr := sg.leastFrequentVertex(indices)
	// 	errors.Logf("DEBUG", "leastFr %v adj %v freq %v label %v", leastFr, sg.Adj[leastFr], indices.G.ColorFrequency(sg.V[leastFr].Color), indices.G.Colors[sg.V[leastFr].Color])
	// 	most := sg.mostConnected()
	// 	errors.Logf("DEBUG", "most %v adj %v freq %v label %v", most, sg.Adj[most], indices.G.ColorFrequency(sg.V[most].Color), indices.G.Colors[sg.V[most].Color])
	// 	leastC := sg.leastConnected()[0]
	// 	errors.Logf("DEBUG", "leastC %v adj %v freq %v label %v", leastC, sg.Adj[leastC], indices.G.ColorFrequency(sg.V[leastC].Color), indices.G.Colors[sg.V[leastC].Color])
	// }

	chain := sg.edgeChain(startIdx)
	vembs := sg.startEmbeddings(indices, startIdx)

	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		stack = append(stack, entry{vemb, 0})
	}

	type idxId struct {
		idx, id int
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
				if !emb.Exists(indices.G) {
					errors.Logf("FOUND", "NOT EXISTS\n  builder %v\n    built %v\n  pattern %v", i.ids, emb, emb.SG)
					panic("wat")
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

func (sg *SubGraph) mostConnected() int {
	maxAdj := 0
	maxIdx := 0
	for idx, adj := range sg.Adj {
		if maxAdj < len(adj) {
			maxAdj = len(adj)
			maxIdx = idx
		}
	}
	return maxIdx
}

func (sg *SubGraph) leastConnectedAndFrequent(indices *Indices) int {
	minAdj := 0
	minIdx := -1
	minFreq := -1
	for idx, adj := range sg.Adj {
		f := indices.G.ColorFrequency(sg.V[idx].Color)
		if minIdx < 0 || minAdj > len(adj) || (minAdj == len(adj) && minFreq > f) {
			minAdj = len(adj)
			minFreq = f
			minIdx = idx
		}
	}
	return minIdx
}

func (sg *SubGraph) leastConnected() []int {
	minAdj := 0
	minIdx := make([]int, 0, 10)
	for idx, adj := range sg.Adj {
		if len(minIdx) == 0 || minAdj > len(adj) {
			minAdj = len(adj)
			minIdx = minIdx[:0]
		}
		if minAdj == len(adj) {
			minIdx = append(minIdx, idx)
		}
	}
	return minIdx
}

func (sg *SubGraph) leastConnectedAndExts(indices *Indices) int {
	c := sg.leastConnected()
	if len(c) == 1 {
		return c[0]
	}
	minExts := -1
	minIdx := -1
	for _, idx := range c {
		exts := sg.extensionsFrom(indices, idx)
		if minIdx == -1 || minExts > exts {
			minExts = exts
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
	}
	return embs
}

// this is really a breadth first search from the given idx
func (sg *SubGraph) edgeChain(startIdx int) []*Edge {
	if startIdx >= len(sg.V) {
		panic("startIdx out of range")
	}
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
			if len(sg.Adj[sg.E[e].Src]) < len(sg.Adj[sg.E[e].Targ]) {
				queue.EnqueFront(types.Int(sg.E[e].Src))
				queue.EnqueFront(types.Int(sg.E[e].Targ))
			} else {
				queue.EnqueFront(types.Int(sg.E[e].Targ))
				queue.EnqueFront(types.Int(sg.E[e].Src))
			}
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
	srcId, targId := cur.ids(e.Src, e.Targ)
	if srcId == -1 && targId == -1 {
		panic("src and targ == -1. Which means the edge chain was not connected.")
	} else if srcId != -1 && targId != -1 {
		// both src and targ are in the builder so we can just add this edge
		if indices.HasEdge(srcId, targId, e.Color) {
			do(cur)
		}
	} else if srcId != -1 {
		indices.TargsFromSrc(srcId, e.Color, sg.V[e.Targ].Color, cur, func(targId int) {
			do(&IdNode{Id: targId, Idx: e.Targ, Prev: cur})
		})
	} else if targId != -1 {
		indices.SrcsToTarg(targId, e.Color, sg.V[e.Src].Color, cur, func(srcId int) {
			do(&IdNode{Id: srcId, Idx: e.Src, Prev: cur})
		})
	} else {
		panic("unreachable")
	}
}

func (sg *SubGraph) extensionsFrom(indices *Indices, idx int) int {
	total := 0
	for _, eid := range sg.Adj[idx] {
		e := &sg.E[eid]
		for _, id := range indices.ColorIndex[sg.V[idx].Color] {
			sg.extendEmbedding(indices, &IdNode{Id: id, Idx: idx}, e, func(_ *IdNode) {
				total++
			})
		}
	}
	return total
}

