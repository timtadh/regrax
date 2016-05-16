package subgraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/linked"
	// "github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

// Tim You are Here
// TODO: Construct the supported Embeddings of the Overlap.
// I then need to port ../extensions.go to use the Overlap graph.

type Overlap struct {
	SG  *SubGraph
	Ids []types.Set // the embeddings for each vertex
}

func (sg *SubGraph) FindVertexEmbeddings(indices *Indices, minSupport int) *Overlap {
	startIdx := 0
	chain := sg.edgeChain(indices, nil, startIdx)
	b := BuildOverlap(len(sg.V), len(sg.E)).Fillable().Ctx(func(b *FillableOverlapBuilder) {
		b.SetVertex(startIdx, sg.V[0].Color, indices.IdSet(sg.V[0].Color))
	})
	for _, e := range chain {
		// errors.Logf("VE-DEBUG", "edge %v", e)
		unsupported := sg.pruneVertices(minSupport, indices, b, sg.extendOverlap(indices, b, e))
		if unsupported {
			return nil
		}
		// errors.Logf("VE-DEBUG", "so far %v", b)
	}
	return b.Build()
}

func (o *Overlap) SupportedEmbeddings(indices *Indices) []*Embedding {
	type entry struct {
		b *FillableEmbeddingBuilder
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	support, idxs := o.MinSupported()
	startIdx := idxs[0]
	chain := o.SG.edgeChain(indices, nil, startIdx)
	embs := make([]*Embedding, 0, support)
	for x, next := o.Ids[startIdx].Items()(); next != nil; x, next = next() {
		id := int(x.(types.Int))
		color := o.SG.V[startIdx].Color
		b := BuildEmbedding(len(o.SG.V), len(o.SG.E)).Fillable().
				Ctx(func(b *FillableEmbeddingBuilder) {
					b.SetVertex(startIdx, color, id)
				})
		stack := make([]entry, 0, len(chain))
		stack = append(stack, entry{b, 0})
		for len(stack) > 0 {
			var item entry
			item, stack = pop(stack)
			if item.eid >= len(chain) {
				embs = append(embs, item.b.Build())
				break
			}
			// b := item.b
			// e := chain[item.eid]
			// exts, addedIdx := o.SG.extendEmbedding(indices, b, e)
			// for _, ext := range exts {
			// 	if addedIdx < 0 || o.Ids[addedIdx].Has(types.Int(ext.Ids[addedIdx])) {
			// 		stack = append(stack, entry{ext, item.eid+1})
			// 	}
			// }
		}
	}
	return embs
}

func (o *Overlap) MinSupported() (support int, vIdxs []int) {
	idxs := make([]int, 0, len(o.Ids))
	min := -1
	for idx, ids := range o.Ids {
		if min < 0 || min > ids.Size() {
			min = ids.Size()
			idxs = idxs[:0]
		}
		if min == ids.Size() {
			idxs = append(idxs, idx)
		}
	}
	return min, idxs
}

func (sg *SubGraph) extendOverlap(indices *Indices, b *FillableOverlapBuilder, e *Edge) (dirty *linked.UniqueDeque) {
	/*
	src := b.V[e.Src].Idx
	targ := b.V[e.Targ].Idx

	
	if src == -1 && targ == -1 {
		panic("src and targ == -1. Which means the edge chain was not connected.")
	} else if src != -1 && targ != -1 {
		b.AddEdge(&b.V[e.Src], &b.V[e.Targ], e.Color)
	} else if src != -1 {
		targs := set.NewSortedSet(10)
		for srcId, next := b.Ids[src].Items()(); next != nil; srcId, next = next() {
			indices.TargsFromSrc(int(srcId.(types.Int)), e.Color, sg.V[e.Targ].Color, nil, func(targ int) {
				targs.Add(types.Int(targ))
			})
		}
		b.SetVertex(e.Targ, sg.V[e.Targ].Color, targs)
		b.AddEdge(&b.V[e.Src], &b.V[e.Targ], e.Color)
	} else if targ != -1 {
		srcs := set.NewSortedSet(10)
		for targId, next := b.Ids[targ].Items()(); next != nil; targId, next = next() {
			indices.SrcsToTarg(int(targId.(types.Int)), e.Color, sg.V[e.Src].Color, nil, func(src int) {
				srcs.Add(types.Int(src))
			})
		}
		b.SetVertex(e.Src, sg.V[e.Src].Color, srcs)
		b.AddEdge(&b.V[e.Src], &b.V[e.Targ], e.Color)
	} else {
		panic("unreachable")
	}
	*/
	dirty = linked.NewUniqueDeque()
	dirty.Push(types.Int(e.Src))
	dirty.Push(types.Int(e.Targ))
	return dirty
}

func (sg *SubGraph) pruneVertices(minSupport int, indices *Indices, b *FillableOverlapBuilder, dirty *linked.UniqueDeque) (unsup bool) {
	for dirty.Size() > 0 {
		idx, err := dirty.DequeBack()
		if err != nil {
			panic(errors.Errorf("should not be possible").(*errors.Error).Chain(err))
		}
		unsup = sg.pruneVertex(int(idx.(types.Int)), minSupport, indices, b, dirty)
		if unsup {
			return true
		}
	}
	return false
}

func (sg *SubGraph) pruneVertex(idx, minSupport int, indices *Indices, b *FillableOverlapBuilder, dirty *linked.UniqueDeque) (unsup bool) {
	// todo next
	if b.Ids[idx].Size() < minSupport {
		return true
	}
	changed := false
	for id, next := b.Ids[idx].Copy().Items()(); next != nil; id, next = next() {
		if !sg.hasEveryEdge(idx, int(id.(types.Int)), indices, b) {
			b.Ids[idx].Delete(id)
			changed = true
		}
		if b.Ids[idx].Size() < minSupport {
			return true
		}
	}
	if changed {
		for _, eidx := range b.Adj[idx] {
			e := &b.E[eidx]
			if e.Src != idx {
				dirty.EnqueFront(types.Int(e.Src))
			}
			if e.Targ != idx {
				dirty.EnqueFront(types.Int(e.Targ))
			}
		}
	}
	return false
}

func (sg *SubGraph) hasEveryEdge(idx, id int, indices *Indices, b *FillableOverlapBuilder) (bool) {
	for _, eidx := range b.Adj[idx] {
		e := &b.E[eidx]
		found := false
		if e.Src == idx {
			srcId := id
			for tid, next := b.Ids[e.Targ].Items()(); next != nil; tid, next = next() {
				targId := int(tid.(types.Int))
				if indices.HasEdge(srcId, targId, e.Color) {
					found = true
					break
				}
			}
		} else {
			targId := id
			for sid, next := b.Ids[e.Src].Items()(); next != nil; sid, next = next() {
				srcId := int(sid.(types.Int))
				if indices.HasEdge(srcId, targId, e.Color) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (o *Overlap) String() string {
	V := make([]string, 0, len(o.SG.V))
	E := make([]string, 0, len(o.SG.E))
	for i := range o.SG.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			o.SG.V[i].Color,
			o.Ids[i],
		))
	}
	for _, e := range o.SG.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			e.Color,
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(o.SG.E), len(o.SG.V), strings.Join(V, ""), strings.Join(E, ""))
}

func (o *Overlap) Pretty(colors []string) string {
	V := make([]string, 0, len(o.SG.V))
	E := make([]string, 0, len(o.SG.E))
	for i, v := range o.SG.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			colors[v.Color],
			o.Ids[i],
		))
	}
	for _, e := range o.SG.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			colors[e.Color],
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(o.SG.E), len(o.SG.V), strings.Join(V, ""), strings.Join(E, ""))
}

