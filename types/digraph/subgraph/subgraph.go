package subgraph

import (
	"fmt"
	"encoding/binary"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/types/digraph/ext"
	"github.com/timtadh/sfp/stores/int_int"
)


type SubGraph struct {
	V []Vertex
	E []Edge
	Adj [][]int
}

type Vertex struct {
	Idx int
	Color int
}

type Edge struct {
	Src, Targ, Color int
}

type EmbIterator func()(*goiso.SubGraph, EmbIterator)

func EmptySubGraph() *SubGraph {
	return &SubGraph{}
}

// Note since *SubGraphs are constructed from *goiso.SubGraphs they are in
// canonical ordering. This is a necessary assumption for Embeddings() to 
// work properly.
func NewSubGraph(sg *goiso.SubGraph) *SubGraph {
	if sg == nil {
		return &SubGraph{
			V: make([]Vertex, 0),
			E: make([]Edge, 0),
			Adj: make([][]int, 0),
		}
	}
	pat := &SubGraph{
		V: make([]Vertex, len(sg.V)),
		E: make([]Edge, len(sg.E)),
		Adj: make([][]int, len(sg.V)),
	}
	for i := range sg.V {
		pat.V[i].Idx = i
		pat.V[i].Color = sg.V[i].Color
		pat.Adj[i] = make([]int, 0, 5)
	}
	for i := range sg.E {
		pat.E[i].Src = sg.E[i].Src
		pat.E[i].Targ = sg.E[i].Targ
		pat.E[i].Color = sg.E[i].Color
		pat.Adj[pat.E[i].Src] = append(pat.Adj[pat.E[i].Src], i)
		pat.Adj[pat.E[i].Targ] = append(pat.Adj[pat.E[i].Targ], i)
	}
	return pat
}

func LoadSubgraphFromLabel(label []byte) (*SubGraph, error) {
	sg := new(SubGraph)
	err := sg.UnmarshalBinary(label)
	if err != nil {
		return nil, err
	}
	return sg, nil
}

func (sg *SubGraph) Embeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender) ([]*goiso.SubGraph, error) {
	embeddings := make([]*goiso.SubGraph, 0, 10)
	err := sg.DoEmbeddings(G, ColorMap, extender, func(emb *goiso.SubGraph) error {
		embeddings = append(embeddings, emb)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return embeddings, nil

}

func (sg *SubGraph) DoEmbeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender, do func(*goiso.SubGraph) error) error {
	ei, err := sg.IterEmbeddings(G, ColorMap, extender)
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


func (sg *SubGraph) IterEmbeddings(G *goiso.Graph, ColorMap int_int.MultiMap, extender *ext.Extender) (ei EmbIterator, err error) {
	type entry struct {
		emb *goiso.SubGraph
		eid int
	}
	pop := func(stack []entry) (entry, []entry) {
		return stack[len(stack)-1], stack[0:len(stack)-1]
	}

	if len(sg.V) == 0 {
		return nil, nil
	}
	startIdx := sg.LeastCommonVertex(G)
	chain := sg.EdgeChainFrom(startIdx)
	vembs, err := sg.VertexEmbeddings(G, ColorMap, startIdx)
	if err != nil {
		return nil, err
	}

	seen := hashtable.NewLinearHash()
	stack := make([]entry, 0, len(vembs)*2)
	for _, vemb := range vembs {
		stack = append(stack, entry{vemb, 0})
	}

	ei = func() (*goiso.SubGraph, EmbIterator) {
		for len(stack) > 0 {
			var i entry
			i, stack = pop(stack)
			label := types.ByteSlice(i.emb.Serialize())
			if seen.Has(label) {
				continue
			}
			seen.Put(label, nil)
			// otherwise success we have an embedding we haven't seen
			if i.eid >= len(chain) {
				if sg.Matches(i.emb) {
					// sweet we can yield this embedding!
					return i.emb, ei
				}
				// nope wasn't an embedding drop it
			} else {
				// ok extend the embedding
				for _, ext := range sg.ExtendEmbedding(G, extender, i.emb, chain[i.eid]) {
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
	exts := ext.NewCollector(-1)
	srcs := sg.findSrcs(cur, e)
	// errors.Logf("DEBUG", "  srcs %v", srcs)
	seen := make(map[int]bool)
	added := 0
	for _, src := range srcs {
		for _, ke := range sg.findEdgesFromSrc(G, cur, src, e) {
			// errors.Logf("DEBUG", "    ke %v %v", ke.Idx, ke)
			if !seen[ke.Idx] {
				seen[ke.Idx] = true
				extender.Extend(cur, ke, exts.Ch())
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
				extender.Extend(cur, pe, exts.Ch())
				added += 1
			}
		}
	}
	exts.Wait(added)
	return exts.Collection()
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

func (sg *SubGraph) MarshalBinary() ([]byte, error) {
	return sg.Label(), nil
}

func (sg *SubGraph) UnmarshalBinary(bytes []byte) (error) {
	if sg.V != nil || sg.E != nil || sg.Adj != nil {
		return errors.Errorf("sg is already loaded! will not load serialized data")
	}
	if len(bytes) < 8 {
		return errors.Errorf("bytes was too small %v < 8", len(bytes))
	}
	lenE := int(binary.BigEndian.Uint32(bytes[0:4]))
	lenV := int(binary.BigEndian.Uint32(bytes[4:8]))
	off := 8
	expected := 8 + lenV*4 + lenE*12
	if len(bytes) < expected {
		return errors.Errorf("bytes was too small %v < %v", len(bytes), expected)
	}
	sg.V = make([]Vertex, lenV)
	sg.E = make([]Edge, lenE)
	sg.Adj = make([][]int, lenV)
	for i := 0; i < lenV; i++ {
		s := off + i*4
		e := s + 4
		color := int(binary.BigEndian.Uint32(bytes[s:e]))
		sg.V[i].Idx = i
		sg.V[i].Color = color
		sg.Adj[i] = make([]int, 0, 5)
	}
	off += lenV*4
	for i := 0; i < lenE; i++ {
		s := off + i*12
		e := s + 4
		src := int(binary.BigEndian.Uint32(bytes[s:e]))
		s += 4
		e += 4
		targ := int(binary.BigEndian.Uint32(bytes[s:e]))
		s += 4
		e += 4
		color := int(binary.BigEndian.Uint32(bytes[s:e]))
		sg.E[i].Src = src
		sg.E[i].Targ = targ
		sg.E[i].Color = color
		sg.Adj[src] = append(sg.Adj[src], i)
		sg.Adj[targ] = append(sg.Adj[targ], i)
	}
	return nil
}

func (sg *SubGraph) Label() []byte {
	size := 8 + len(sg.V)*4 + len(sg.E)*12
	label := make([]byte, size)
	binary.BigEndian.PutUint32(label[0:4], uint32(len(sg.E)))
	binary.BigEndian.PutUint32(label[4:8], uint32(len(sg.V)))
	off := 8
	for i, v := range sg.V {
		s := off + i*4
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(v.Color))
	}
	off += len(sg.V)*4
	for i, edge := range sg.E {
		s := off + i*12
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Src))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Targ))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Color))
	}
	return label
}

func (sg *SubGraph) String() string {
	V := make([]string, 0, len(sg.V))
	E := make([]string, 0, len(sg.E))
	for _, v := range sg.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			v.Idx,
			v.Color,
		))
	}
	for _, e := range sg.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			e.Color,
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(sg.E), len(sg.V), strings.Join(V, ""), strings.Join(E, ""))
}



