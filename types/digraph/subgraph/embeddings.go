package subgraph

import ()

import (
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/types/digraph/ext"
)




func (sg *SubGraph) Embeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender) ([]*goiso.SubGraph, error) {
	embeddings := make([]*goiso.SubGraph, 0, 10)
	err := sg.DoEmbeddings(G, ColorMap, extender, nil, func(emb *goiso.SubGraph) error {
		embeddings = append(embeddings, emb)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return embeddings, nil

}

func (sg *SubGraph) DoEmbeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender, pruner Pruner, do func(*goiso.SubGraph) error) error {
	ei, err := sg.IterEmbeddings(G, ColorMap, extender, pruner)
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

func (sg *SubGraph) IterEmbeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender, pruner Pruner) (ei EmbIterator, err error) {
	type entry struct {
		emb *goiso.SubGraph
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0 : len(stack)-1]
	}
	vertexSet := func(emb *goiso.SubGraph) *set.SortedSet {
		vertices := set.NewSortedSet(len(emb.V))
		for i := range emb.V {
			vertices.Add(types.Int(emb.V[i].Id))
		}
		return vertices
	}

	if len(sg.V) == 0 {
		ei = func() (*goiso.SubGraph, EmbIterator) {
			return nil, nil
		}
		return ei, nil
	}
	startIdx := sg.LeastCommonVertex(G)
	chain := sg.EdgeChainFrom(startIdx)
	vembs, err := sg.VertexEmbeddings(G, ColorMap, startIdx)
	if err != nil {
		return nil, err
	}

	var prune func(*goiso.SubGraph) bool = nil
	if pruner != nil {
		prune = pruner(startIdx, chain)
	}

	seen := hashtable.NewLinearHash()
	visited := hashtable.NewLinearHash()
	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		stack = append(stack, entry{vemb, 0})
	}

	ei = func() (*goiso.SubGraph, EmbIterator) {
		for len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
			if prune != nil && prune(i.emb) {
				continue
			}
			// otherwise success we have an embedding we haven't seen
			if i.eid >= len(chain) {
				// check that this is the subgraph we sought
				if sg.Matches(i.emb) {
					// check for automorphisms
					vertices := vertexSet(i.emb)
					if !visited.Has(vertices) {
						visited.Put(vertices, nil)
						// sweet we can yield this embedding!
						return i.emb, ei
					}
				}
				// nope wasn't an embedding drop it
			} else {
				// ok extend the embedding
				for _, ext := range sg.ExtendEmbedding(G, extender, i.emb, chain[i.eid]) {
					label := types.ByteSlice(ext.Serialize())
					if seen.Has(label) {
						continue
					}
					seen.Put(label, nil)
					stack = append(stack, entry{ext, i.eid + 1})
				}
			}
		}
		return nil, nil
	}
	return ei, nil
}

func (sg *SubGraph) Matches(emb *goiso.SubGraph) bool {
	if len(sg.V) != len(emb.V) {
		return false
	}
	if len(sg.E) != len(emb.E) {
		return false
	}
	for i := range sg.V {
		if sg.V[i].Color != emb.V[i].Color {
			return false
		}
	}
	for i := range sg.E {
		if sg.E[i].Src != emb.E[i].Src {
			return false
		}
		if sg.E[i].Targ != emb.E[i].Targ {
			return false
		}
		if sg.E[i].Color != emb.E[i].Color {
			return false
		}
	}
	return true
}

func (sg *SubGraph) LeastCommonVertex(G *goiso.Graph) int {
	minFreq := -1
	minIdx := -1
	for i := range sg.V {
		f := G.ColorFrequency(sg.V[i].Color)
		if f < minFreq || minIdx == -1 {
			minFreq = f
			minIdx = i
		}
	}
	return minIdx
}

func (sg *SubGraph) VertexEmbeddings(G *goiso.Graph, ColorMap int_int.MultiMap, idx int) ([]*goiso.SubGraph, error) {
	embs := make([]*goiso.SubGraph, 0, G.ColorFrequency(sg.V[idx].Color))
	err := ColorMap.DoFind(int32(sg.V[idx].Color), func(color, gIdx int32) error {
		sg, _ := G.VertexSubGraph(int(gIdx))
		embs = append(embs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return embs, nil
}

// this is really a depth first search from the given idx
func (sg *SubGraph) EdgeChainFrom(idx int) []*Edge {
	edges := make([]*Edge, 0, len(sg.E))
	added := make(map[int]bool, len(sg.E))
	seen := make(map[int]bool, len(sg.V))
	var visit func(int)
	visit = func(u int) {
		seen[u] = true
		for _, e := range sg.Adj[u] {
			if !added[e] {
				added[e] = true
				edges = append(edges, &sg.E[e])
			}
		}
		for _, e := range sg.Adj[u] {
			// errors.Logf("DEBUG", "u %v adj %v", u, sg.E[e])
			v := sg.E[e].Src
			if !seen[v] {
				visit(v)
			}
			v = sg.E[e].Targ
			if !seen[v] {
				visit(v)
			}
		}
	}
	visit(idx)
	// errors.Logf("DEBUG", "edge chain seen %v", seen)
	// errors.Logf("DEBUG", "edge chain added %v", added)
	return edges
}

func (sg *SubGraph) ExtendEmbedding(G *goiso.Graph, extender *ext.Extender, cur *goiso.SubGraph, e *Edge) []*goiso.SubGraph {
	// errors.Logf("DEBUG", "extend emb %v with %v", cur.Label(), e)
	// exts := ext.NewCollector(-1)
	exts := make([]*goiso.SubGraph, 0, 10)
	srcs := sg.findSrcs(cur, e)
	// errors.Logf("DEBUG", "  srcs %v", srcs)
	seen := make(map[int]bool)
	added := 0
	for _, src := range srcs {
		for _, ke := range sg.findEdgesFromSrc(G, cur, src, e) {
			// errors.Logf("DEBUG", "    ke %v %v", ke.Idx, ke)
			if !seen[ke.Idx] {
				seen[ke.Idx] = true
				// extender.Extend(cur, ke, exts.Ch())
				x, _ := cur.EdgeExtend(ke)
				exts = append(exts, x)
				added += 1
			}
		}
	}
	targs := sg.findTargs(cur, e)
	// errors.Logf("DEBUG", "  targs %v", targs)
	for _, targ := range targs {
		for _, pe := range sg.findEdgesFromTarg(G, cur, targ, e) {
			// errors.Logf("DEBUG", "    pe %v %v", pe.Idx, pe)
			if !seen[pe.Idx] {
				seen[pe.Idx] = true
				// extender.Extend(cur, pe, exts.Ch())
				x, _ := cur.EdgeExtend(pe)
				exts = append(exts, x)
				added += 1
			}
		}
	}
	// return exts.Wait(added)
	return exts
}

func (sg *SubGraph) findSrcs(cur *goiso.SubGraph, e *Edge) []int {
	color := sg.V[e.Src].Color
	srcs := make([]int, 0, 10)
	for i := range cur.V {
		if cur.V[i].Color == color {
			srcs = append(srcs, i)
		}
	}
	return srcs
}

func (sg *SubGraph) findTargs(cur *goiso.SubGraph, e *Edge) []int {
	color := sg.V[e.Targ].Color
	targs := make([]int, 0, 10)
	for i := range cur.V {
		if cur.V[i].Color == color {
			targs = append(targs, i)
		}
	}
	return targs
}

func (sg *SubGraph) findEdgesFromSrc(G *goiso.Graph, cur *goiso.SubGraph, src int, e *Edge) []*goiso.Edge {
	srcDtIdx := cur.V[src].Id
	tcolor := sg.V[e.Targ].Color
	ecolor := e.Color
	edges := make([]*goiso.Edge, 0, 10)
	for _, ke := range G.Kids[srcDtIdx] {
		if ke.Color != ecolor {
			continue
		} else if G.V[ke.Targ].Color != tcolor {
			continue
		}
		if !cur.HasEdge(goiso.ColoredArc{ke.Arc, ke.Color}) {
			edges = append(edges, ke)
		}
	}
	return edges
}

func (sg *SubGraph) findEdgesFromTarg(G *goiso.Graph, cur *goiso.SubGraph, targ int, e *Edge) []*goiso.Edge {
	targDtIdx := cur.V[targ].Id
	scolor := sg.V[e.Src].Color
	ecolor := e.Color
	edges := make([]*goiso.Edge, 0, 10)
	for _, pe := range G.Parents[targDtIdx] {
		if pe.Color != ecolor {
			continue
		} else if G.V[pe.Src].Color != scolor {
			continue
		}
		if !cur.HasEdge(goiso.ColoredArc{pe.Arc, pe.Color}) {
			edges = append(edges, pe)
		}
	}
	return edges
}


