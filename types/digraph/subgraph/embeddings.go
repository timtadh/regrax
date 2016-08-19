package subgraph

import (
	"fmt"
	// "math/rand"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/heap"
	"github.com/timtadh/data-structures/list"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
// "github.com/timtadh/sfp/stats"
)

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
	ei, _, err := sg.IterEmbeddings(indices, nil, nil, nil)
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

func FilterAutomorphs(it EmbIterator, dropped *[]*Embedding, err error) (ei EmbIterator, _ *[]*Embedding, _ error) {
	if err != nil {
		return nil, nil, err
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
	return ei, dropped,  nil
}

type IdNode struct {
	Id   int
	Idx  int
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

func (ids *IdNode) addOrReplace(id, idx int) *IdNode {
	for c := ids; c != nil; c = c.Prev {
		if idx == c.Idx {
			c.Id = id
			return ids
		}
	}
	return &IdNode{Id: id, Idx: idx, Prev: ids}
}

func (sg *SubGraph) IterEmbeddings(indices *Indices, prunePoints types.Set, overlap []map[int]bool, prune func(*IdNode) bool) (ei EmbIterator, unsup *[]*Embedding, err error) {
	if len(sg.V) == 0 {
		ei = func() (*Embedding, EmbIterator) {
			return nil, nil
		}
		return ei, nil, nil
	}
	type entry struct {
		ids *IdNode
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	dropped := make([]*Embedding, 0, 10)
	startIdx := sg.leastExts(indices, overlap)
	chain := sg.edgeChain(indices, overlap, startIdx)
	vembs := sg.startEmbeddings(indices, startIdx)
	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		stack = append(stack, entry{vemb, 0})
	}

	ei = func() (*Embedding, EmbIterator) {
		for len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
			if prune != nil && prune(i.ids) {
				continue
			}
			// otherwise success we have an embedding we haven't seen
			if i.eid >= len(chain) {
				// check that this is the subgraph we sought
				emb := &Embedding{
					SG:  sg,
					Ids: i.ids.list(len(sg.V)),
				}
				if !emb.Exists(indices.G) {
					errors.Logf("FOUND", "NOT EXISTS\n  builder %v\n    built %v\n  pattern %v", i.ids, emb, emb.SG)
					panic("wat")
				}
				if sg.Equals(emb) {
					// sweet we can yield this embedding!
					return emb, ei
				} else {
					// nope wasn't an embedding drop it
					dropped = append(dropped, emb)
				}
			} else {
				var part *Embedding = nil
				if prunePoints != nil || len(sg.E) < 2 {
				// errors.Logf("DEBUG", "prunePoints %v", prunePoints)
					part = sg.partial(chain[:i.eid], i.ids)
					if prunePoints != nil && prunePoints.Has(part) {
						// errors.Logf("DEBUG", "pruned %v", part)
						continue
					}
				}
				// ok extend the embedding
				size := len(stack)
				sg.extendEmbedding(indices, i.ids, &sg.E[chain[i.eid]], overlap, func(ext *IdNode) {
					stack = append(stack, entry{ext, i.eid + 1})
				})
				if size == len(stack) && part != nil {
					// errors.Logf("DEBUG", "dropping %v", part)
					dropped = append(dropped, part)
				}
			}
		}
		// errors.Logf("DEBUG", "dropped %v", dropped)
		return nil, nil
	}
	return ei, &dropped, nil
}


func (sg *SubGraph) partial(edgeChain []int, ids *IdNode) *Embedding {
	b := BuildEmbedding(len(sg.V), len(edgeChain))
	vidxs := make(map[int]*Vertex)
	addVertex := func(vidx, color, vid int) {
		if _, has := vidxs[vidx]; !has {
			vidxs[vidx] = b.AddVertex(color, vid)
		} else {
			panic("double add")
		}
	}
	for c := ids; c != nil; c = c.Prev {
		addVertex(c.Idx, sg.V[c.Idx].Color, c.Id)
	}
	for _, eid := range edgeChain {
		s := sg.E[eid].Src
		t := sg.E[eid].Targ
		color := sg.E[eid].Color
		b.AddEdge(vidxs[s], vidxs[t], color)
	}
	return b.Build()
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

func (sg *SubGraph) mostCard(indices *Indices) int {
	maxCard := 0
	maxIdx := 0
	for idx := range sg.V {
		card := sg.vertexCard(indices, idx)
		if maxCard < card {
			maxCard = card
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

func (sg *SubGraph) leastConnectedAndExts(indices *Indices, overlap []map[int]bool) int {
	c := sg.leastConnected()
	if len(c) == 1 {
		return c[0]
	}
	minExts := -1
	minIdx := -1
	for _, idx := range c {
		exts := sg.extensionsFrom(indices, overlap, idx)
		if minIdx == -1 || minExts > exts {
			minExts = exts
			minIdx = idx
		}
	}
	return minIdx
}

func (sg *SubGraph) leastExts(indices *Indices, overlap []map[int]bool) int {
	minExts := -1
	minFreq := -1
	minIdx := -1
	for idx := range sg.V {
		freq := indices.G.ColorFrequency(sg.V[idx].Color)
		exts := sg.extensionsFrom(indices, overlap, idx)
		if minIdx == -1 || minExts > exts || (minExts == exts && minFreq > freq) {
			minExts = exts
			minFreq = freq
			minIdx = idx
		}
	}
	return minIdx
}

func (sg *SubGraph) mostExts(indices *Indices, overlap []map[int]bool) int {
	maxExts := -1
	maxIdx := -1
	for idx := range sg.V {
		exts := sg.extensionsFrom(indices, overlap, idx)
		if maxIdx == -1 || maxExts < exts {
			maxExts = exts
			maxIdx = idx
		}
	}
	return maxIdx
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
func (sg *SubGraph) edgeChain(indices *Indices, overlap []map[int]bool, startIdx int) []int {
	other := func(u int, e int) int {
		s := sg.E[e].Src
		t := sg.E[e].Targ
		var v int
		if s == u {
			v = t
		} else if t == u {
			v = s
		} else {
			panic("unreachable")
		}
		return v
	}
	if startIdx >= len(sg.V) {
		panic("startIdx out of range")
	}
	colors := make(map[int]bool, len(sg.V))
	edges := make([]int, 0, len(sg.E))
	added := make(map[int]bool, len(sg.E))
	seen := make(map[int]bool, len(sg.V))
	queue := heap.NewUnique(heap.NewMinHeap(len(sg.V)))
	queue.Add(0, types.Int(startIdx))
	prevs := make([]int, 0, len(sg.V))
	for queue.Size() > 0 {
		u := int(queue.Pop().(types.Int))
		if seen[u] {
			continue
		}
	find_edge:
		for i := len(prevs) - 1; i >= 0; i-- {
			prev := prevs[i]
			for _, e := range sg.Adj[prev] {
				v := other(prev, e)
				if v == u {
					if !added[e] {
						edges = append(edges, e)
						added[e] = true
						break find_edge
					}
				}
			}
		}
		// if len(sg.E) > 0 {
		// 	errors.Logf("DEBUG", "vertex %v", u)
		// }
		seen[u] = true
		colors[sg.V[u].Color] = true
		for i, e := range sg.Adj[u] {
			v := other(u, e)
			if seen[v] {
				continue
			}
			p := i
			// p = sg.vertexCard(indices, v)
			// p = indices.G.ColorFrequency(sg.V[v].Color)
			// p = len(sg.Adj[v]) - 1
			extsFrom := sg.extensionsFrom(indices, overlap, v, u)
			p = extsFrom // +  + indices.G.ColorFrequency(sg.V[v].Color)
			// p = extsFrom + len(sg.Adj[v]) - 1 + indices.G.ColorFrequency(sg.V[v].Color)
			if extsFrom == 0 {
				// p = // indices.G.ColorFrequency(sg.V[v].Color) // * sg.vertexCard(indices, v)
				// p = sg.vertexCard(indices, v)
				p = sg.extensionsFrom(indices, overlap, v) * 4 // penalty for all targets being known
			}
			if !colors[v] {
				p /= 2
			}
			for _, aid := range sg.Adj[v] {
				n := other(v, aid)
				if !seen[n] {
					p -= sg.extensionsFrom(indices, overlap, n, v, u)
					// p += sg.vertexCard(indices, n)
					// a := &sg.E[aid]
					// s := sg.V[a.Src].Color
					// t := sg.V[a.Targ].Color
					// p += indices.EdgeCounts[Colors{SrcColor: s, TargColor: t, EdgeColor: a.Color}]
				}
			}
			// if len(sg.E) > 0 {
			// 	errors.Logf("DEBUG", "add p %v vertex %v extsFrom %v", p, v, extsFrom)
			// }
			queue.Add(p, types.Int(v))
		}
		prevs = append(prevs, u)
	}
	for e := range sg.E {
		if !added[e] {
			edges = append(edges, e)
			added[e] = true
		}
	}
	if len(edges) != len(sg.E) {
		panic("assert-fail: len(edges) != len(sg.E)")
	}

	// if len(sg.E) > 0 {
	// 	errors.Logf("DEBUG", "edge chain seen %v", seen)
	// 	errors.Logf("DEBUG", "edge chain added %v", added)
	// 	for _, e := range edges {
	// 		errors.Logf("DEBUG", "edge %v", e)
	// 	}
	// 	// panic("wat")
	// }
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
	idx := len(items) - 1
	for _, item := range items {
		ritems[idx] = item
		idx--
	}
	return "{" + strings.Join(ritems, ", ") + "}"
}

func (sg *SubGraph) extendEmbedding(indices *Indices, cur *IdNode, e *Edge, o []map[int]bool, do func(*IdNode)) {
	doNew := func(newIdx, newId int) {
		if o == nil || len(o[newIdx]) == 0 {
			do(&IdNode{Id: newId, Idx: newIdx, Prev: cur})
		} else if o[newIdx] != nil && o[newIdx][newId] {
			do(&IdNode{Id: newId, Idx: newIdx, Prev: cur})
		}
	}
	srcId, targId := cur.ids(e.Src, e.Targ)
	if srcId == -1 && targId == -1 {
		panic("src and targ == -1. Which means the edge chain was not connected.")
	} else if srcId != -1 && targId != -1 {
		// both src and targ are in the builder so we can just add this edge
		if indices.HasEdge(srcId, targId, e.Color) {
			do(cur)
		}
	} else if srcId != -1 {
		deg := len(sg.Adj[e.Targ])
		indices.TargsFromSrc(srcId, e.Color, sg.V[e.Targ].Color, cur, func(targId int) {
			// TIM YOU ARE HERE:
			// filter these potential targId by ensuring they have adequate degree
			if deg <= indices.Degree(targId) {
				doNew(e.Targ, targId)
			}
		})
	} else if targId != -1 {
		deg := len(sg.Adj[e.Src])
		indices.SrcsToTarg(targId, e.Color, sg.V[e.Src].Color, cur, func(srcId int) {
			// filter these potential srcId by ensuring they have adequate degree
			if deg <= indices.Degree(srcId) {
				doNew(e.Src, srcId)
			}
		})
	} else {
		panic("unreachable")
	}
}

func (sg *SubGraph) extensionsFrom(indices *Indices, overlap []map[int]bool, idx int, excludeIdxs ...int) int {
	total := 0
outer:
	for _, eid := range sg.Adj[idx] {
		e := &sg.E[eid]
		for _, excludeIdx := range excludeIdxs {
			if e.Src == excludeIdx || e.Targ == excludeIdx {
				continue outer
			}
		}
		for _, id := range indices.ColorIndex[sg.V[idx].Color] {
			if overlap == nil || len(overlap[idx]) == 0 || overlap[idx][id] {
				sg.extendEmbedding(indices, &IdNode{Id: id, Idx: idx}, e, overlap, func(_ *IdNode) {
					total++
				})
			}
		}
	}
	return total
}

func (sg *SubGraph) vertexCard(indices *Indices, idx int) int {
	card := 0
	for _, eid := range sg.Adj[idx] {
		e := &sg.E[eid]
		s := sg.V[e.Src].Color
		t := sg.V[e.Targ].Color
		card += indices.EdgeCounts[Colors{SrcColor: s, TargColor: t, EdgeColor: e.Color}]
	}
	return card
}
