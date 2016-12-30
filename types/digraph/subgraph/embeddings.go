package subgraph

import (
	"fmt"
	"math/rand"
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
	"github.com/timtadh/sfp/types/digraph/digraph"
)

type EmbSearchStartPoint uint64

const (
	RandomStart EmbSearchStartPoint = 1 << iota
	LeastFrequent
	MostFrequent
	LeastConnected
	MostConnected
	FewestExtensions
	MostExtensions
	LowestCardinality
	HighestCardinality
)

type EmbIterator func(bool) (*Embedding, EmbIterator)

func (sg *SubGraph) Embeddings(indices *digraph.Indices) ([]*Embedding, error) {
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

func (sg *SubGraph) DoEmbeddings(indices *digraph.Indices, do func(*Embedding) error) error {
	ei := sg.IterEmbeddings(RandomStart, indices, nil, nil)
	for emb, next := ei(false); next != nil; emb, next = next(false) {
		err := do(emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func FilterAutomorphs(it EmbIterator) (ei EmbIterator) {
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
	return ei
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

func (ids *IdNode) hasId(id int) bool {
	for c := ids; c != nil; c = c.Prev {
		if id == c.Id {
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

func (sg *SubGraph) searchStartingPoint(mode EmbSearchStartPoint, indices *digraph.Indices, overlap []map[int]bool) int {
	switch mode {
	case LeastFrequent:
		return argMin(len(sg.V), sg.vertexFrequency(indices))
	case MostFrequent:
		return argMax(len(sg.V), sg.vertexFrequency(indices))
	case LeastConnected:
		return argMin(len(sg.V), sg.vertexConnectedness)
	case MostConnected:
		return argMax(len(sg.V), sg.vertexConnectedness)
	case FewestExtensions:
		return argMin(len(sg.V), sg.vertexExtensions(indices, overlap))
	case MostExtensions:
		return argMax(len(sg.V), sg.vertexExtensions(indices, overlap))
	case LowestCardinality:
		return argMin(len(sg.V), sg.vertexCardinality(indices))
	case HighestCardinality:
		return argMax(len(sg.V), sg.vertexCardinality(indices))
	case RandomStart:
		fallthrough
	default:
		return rand.Intn(len(sg.V))
	}
}

func (sg *SubGraph) IterEmbeddings(spMode EmbSearchStartPoint, indices *digraph.Indices, overlap []map[int]bool, prune func(*IdNode) bool) (ei EmbIterator) {
	if len(sg.V) == 0 {
		ei = func(bool) (*Embedding, EmbIterator) {
			return nil, nil
		}
		return ei
	}
	type entry struct {
		ids *IdNode
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	startIdx := sg.searchStartingPoint(spMode, indices, overlap)
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
		stack = append(stack, entry{vemb, 0})
	}

	pruneLevel := len(chain) + 2

	ei = func(stop bool) (*Embedding, EmbIterator) {
		for !stop && len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
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
				if false {
					if !emb.Exists(indices.G) {
						errors.Logf("FOUND", "NOT EXISTS\n  builder %v\n    built %v\n  pattern %v", i.ids, emb, emb.SG)
						panic("wat")
					}
					if !sg.Equals(emb) {
						errors.Logf("FOUND", "NOT AN EMB\n  builder %v\n    built %v\n  pattern %v", i.ids, emb, emb.SG)
						panic("wat")
					}
				}
				return emb, ei
			} else {
				// ok extend the embedding
				// size := len(stack)
				sg.extendEmbedding(indices, i.ids, &sg.E[chain[i.eid]], overlap, func(ext *IdNode) {
					stack = append(stack, entry{ext, i.eid + 1})
				})
			}
		}
		return nil, nil
	}
	return ei
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

func argMin(length int, f func(int) int) (arg int) {
	min := 0
	arg = -1
	for i := 0; i < length; i++ {
		x := f(i)
		if arg == -1 || x < min {
			min = x
			arg = i
		}
	}
	return arg
}

func argMax(length int, f func(int) int) (arg int) {
	max := 0
	arg = -1
	for i := 0; i < length; i++ {
		x := f(i)
		if arg == -1 || x > max {
			max = x
			arg = i
		}
	}
	return arg
}

func (sg *SubGraph) vertexFrequency(indices *digraph.Indices) func(int) int {
	return func(idx int) int {
		return indices.VertexColorFrequency(sg.V[idx].Color)
	}
}

func (sg *SubGraph) vertexConnectedness(idx int) int {
	return len(sg.Adj[idx])
}

func (sg *SubGraph) vertexCardinality(indices *digraph.Indices) func(int) int {
	return func(idx int) int {
		return sg.vertexCard(indices, idx)
	}
}

func (sg *SubGraph) vertexExtensions(indices *digraph.Indices, overlap []map[int]bool) func(int) int {
	return func(idx int) int {
		return sg.extensionsFrom(indices, overlap, idx)
	}
}

func (sg *SubGraph) startEmbeddings(indices *digraph.Indices, startIdx int) []*IdNode {
	color := sg.V[startIdx].Color
	embs := make([]*IdNode, 0, indices.VertexColorFrequency(color))
	for _, gIdx := range indices.ColorIndex[color] {
		embs = append(embs, &IdNode{VrtEmb:VrtEmb{Id: gIdx, Idx: startIdx}})
	}
	return embs
}

// this is really a breadth first search from the given idx
func (sg *SubGraph) edgeChain(indices *digraph.Indices, overlap []map[int]bool, startIdx int) []int {
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

func (sg *SubGraph) extendEmbedding(indices *digraph.Indices, cur *IdNode, e *Edge, o []map[int]bool, do func(*IdNode)) {
	doNew := func(newIdx, newId int) {
		// enforce forward consistency
		// we need a more performant way to do this
		// for _, xi := range sg.Adj[newIdx] {
		// 	x := &sg.E[xi]
		// 	if x.Src == newIdx {
		// 		if len(indices.SrcIndex[IdColorColor{newId, x.Color, sg.V[x.Targ].Color}]) == 0 {
		// 			return
		// 		}
		// 	} else {
		// 		if len(indices.TargIndex[IdColorColor{newId, x.Color, sg.V[x.Src].Color}]) == 0 {
		// 			return
		// 		}
		// 	}
		// }
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
		outDeg := sg.OutDeg[e.Targ]
		inDeg := sg.InDeg[e.Targ]
		indices.TargsFromSrc(srcId, e.Color, sg.V[e.Targ].Color, cur.hasId, func(targId int) {
			if outDeg <= indices.OutDegree(targId) && inDeg <= indices.InDegree(targId) {
				doNew(e.Targ, targId)
			}
		})
	} else if targId != -1 {
		outDeg := sg.OutDeg[e.Src]
		inDeg := sg.InDeg[e.Src]
		indices.SrcsToTarg(targId, e.Color, sg.V[e.Src].Color, cur.hasId, func(srcId int) {
			if outDeg <= indices.OutDegree(srcId) && inDeg <= indices.InDegree(srcId) {
				doNew(e.Src, srcId)
			}
		})
	} else {
		panic("unreachable")
	}
}

func (sg *SubGraph) extensionsFrom(indices *digraph.Indices, overlap []map[int]bool, idx int, excludeIdxs ...int) int {
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

func (sg *SubGraph) vertexCard(indices *digraph.Indices, idx int) int {
	card := 0
	for _, eid := range sg.Adj[idx] {
		e := &sg.E[eid]
		s := sg.V[e.Src].Color
		t := sg.V[e.Targ].Color
		card += indices.EdgeCounts[digraph.Colors{SrcColor: s, TargColor: t, EdgeColor: e.Color}]
	}
	return card
}
