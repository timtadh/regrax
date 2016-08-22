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

type EmbIterator func(bool) (*Embedding, EmbIterator)

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
	for emb, next := ei(false); next != nil; emb, next = next(false) {
		err := do(emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func FilterAutomorphs(it EmbIterator, dropped *VertexEmbeddings, err error) (ei EmbIterator, _ *VertexEmbeddings, _ error) {
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
	ei = func(stop bool) (emb *Embedding, _ EmbIterator) {
		if it == nil {
			return nil, nil
		}
		for emb, it = it(stop); it != nil; emb, it = it(stop) {
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

type VrtEmb struct {
	Id   int
	Idx  int
}

func (v *VrtEmb) Equals(o types.Equatable) bool {
	a := v
	switch b := o.(type) {
	case *VrtEmb:
		return a.Id == b.Id && a.Idx == b.Idx
	default:
		return false
	}
}

func (v *VrtEmb) Less(o types.Sortable) bool {
	a := v
	switch b := o.(type) {
	case *VrtEmb:
		return a.Id < b.Id || (a.Id == b.Id && a.Idx < b.Idx)
	default:
		return false
	}
}

func (v *VrtEmb) Hash() int {
	return v.Id*3 + v.Idx*5
}

func (v *VrtEmb) Translate(orgLen int, vord []int) *VrtEmb {
	idx := v.Idx
	if idx >= orgLen {
		idx = len(vord) + (idx - orgLen)
	}
	if idx < len(vord) {
		idx = vord[idx]
	}
	return &VrtEmb{
		Idx: idx,
		Id: v.Id,
	}
}

type VertexEmbeddings []*VrtEmb

func (embs VertexEmbeddings) Translate(orgLen int, vord []int) (VertexEmbeddings) {
	translated := make(VertexEmbeddings, len(embs))
	for i := range embs {
		translated[i] = embs[i].Translate(orgLen, vord)
	}
	return translated
}

func (embs VertexEmbeddings) Set() map[VrtEmb]bool {
	s := make(map[VrtEmb]bool, len(embs))
	for _, emb := range embs {
		s[*emb] = true
	}
	return s
}

type IdNode struct {
	VrtEmb
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
	return &IdNode{VrtEmb:VrtEmb{Id: id, Idx: idx}, Prev: ids}
}

func (sg *SubGraph) IterEmbeddings(indices *Indices, prunePoints map[VrtEmb]bool, overlap []map[int]bool, prune func(*IdNode) bool) (ei EmbIterator, unsup *VertexEmbeddings, err error) {
	if len(sg.V) == 0 {
		ei = func(bool) (*Embedding, EmbIterator) {
			return nil, nil
		}
		return ei, nil, nil
	}
	type entry struct {
		ids *IdNode
		eid int
		proc bool
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	dropped := make(VertexEmbeddings, 0, 10)
	//startIdx := sg.mostExts(indices, overlap)
	//startIdx := sg.leastExts(indices, overlap)
	//startIdx := sg.leastConnectedAndExts(indices, overlap)
	startIdx := sg.mostConnected()
	chain := sg.edgeChain(indices, overlap, startIdx)
	seen := make([]map[VrtEmb]bool, len(chain))
	for i := range seen {
		seen[i] = make(map[VrtEmb]bool)
	}
	used := make(map[VrtEmb]bool)
	vembs := sg.startEmbeddings(indices, startIdx)
	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		// seen[vemb.VrtEmb] = true
		stack = append(stack, entry{vemb, 0, false})
	}

	if prunePoints != nil && len(prunePoints) > 0 {
	//	errors.Logf("DEBUG", "prune points %v", set.SortedFromSet(prunePoints))
	}
	pruneLevel := len(chain) + 2

	ei = func(stop bool) (*Embedding, EmbIterator) {
		for !stop && len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
			if !i.proc {
				stack = append(stack, entry{i.ids, i.eid, true})
			} else {
				if prunePoints != nil && i.eid < len(chain) && i.eid <= pruneLevel { //&& i.eid < 3 {
					if !used[i.ids.VrtEmb] {
						seen[i.eid][i.ids.VrtEmb] = true
						// dropped = append(dropped, &i.ids.VrtEmb)
					}
				}
				continue
			}
			if prunePoints != nil && i.eid < pruneLevel && prunePoints[i.ids.VrtEmb] {
				pruneLevel = i.eid
				continue
			}
			if prune != nil && prune(i.ids) {
				if i.eid < pruneLevel {
					pruneLevel = i.eid
				}
				continue
			}
			// otherwise success we have an embedding we haven't seen
			if i.eid >= len(chain) {
				// check that this is the subgraph we sought
				emb := &Embedding{
					SG:  sg,
					Ids: i.ids.list(len(sg.V)),
				}
				for c := i.ids; c != nil; c = c.Prev {
					used[c.VrtEmb] = true
				}
				if !emb.Exists(indices.G) {
					errors.Logf("FOUND", "NOT EXISTS\n  builder %v\n    built %v\n  pattern %v", i.ids, emb, emb.SG)
					panic("wat")
				}
				if sg.Equals(emb) {
					// sweet we can yield this embedding!
					return emb, ei
				}
				// nope wasn't an embedding drop it
				// this should never happen
			} else {
				// ok extend the embedding
				// size := len(stack)
				sg.extendEmbedding(indices, i.ids, &sg.E[chain[i.eid]], overlap, func(ext *IdNode) {
					stack = append(stack, entry{ext, i.eid + 1, false})
				})
				// if size == len(stack) && prunePoints != nil && i.ids.Prev == nil {
					// dropped = append(dropped, &i.ids.VrtEmb)
				// }
			}
		}
		if prunePoints != nil {
			for i := 0; i < pruneLevel && i < len(seen); i++ {
				for ve := range seen[i] {
					if !used[ve] {
						dropped = append(dropped, &ve)
					}
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
		embs = append(embs, &IdNode{VrtEmb:VrtEmb{Id: gIdx, Idx: startIdx}})
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
			do(&IdNode{VrtEmb:VrtEmb{Id: newId, Idx: newIdx}, Prev: cur})
		} else if o[newIdx] != nil && o[newIdx][newId] {
			do(&IdNode{VrtEmb:VrtEmb{Id: newId, Idx: newIdx}, Prev: cur})
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
				sg.extendEmbedding(indices, &IdNode{VrtEmb:VrtEmb{Id: id, Idx: idx}}, e, overlap, func(_ *IdNode) {
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
