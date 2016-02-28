package digraph

import (
	"fmt"
	"encoding/binary"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
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

func (sg *SubGraph) Embeddings(dt *Digraph) ([]*goiso.SubGraph, error) {
	// errors.Logf("DEBUG", "Embeddings of %v", sg)
	if len(sg.V) == 0 {
		return nil, nil
	}
	startIdx := sg.LeastCommonVertex(dt)
	chain := sg.EdgeChainFrom(startIdx)
	cur, err := sg.VertexEmbeddings(dt, startIdx)
	if err != nil {
		return nil, err
	}
	// errors.Logf("DEBUG", "startIdx %v", startIdx)
	// errors.Logf("DEBUG", "chain %v", chain)
	for _, e := range chain {
		// errors.Logf("DEBUG", "cur %v e %v", len(cur), e)
		next := make([]*goiso.SubGraph, 0, len(cur))
		for _, emb := range cur {
			for _, ext := range sg.ExtendEmbedding(dt, emb, e) {
				next = append(next, ext)
			}
		}
		cur = DedupSupported(next)
	}
	final := make([]*goiso.SubGraph, 0, len(cur))
	for _, emb := range cur {
		if sg.Matches(emb) {
			final = append(final, emb)
		}
	}
	return final, nil
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

func (sg *SubGraph) LeastCommonVertex(dt *Digraph) int {
	minFreq := -1
	minIdx := -1
	for i := range sg.V {
		f := dt.G.ColorFrequency(sg.V[i].Color)
		if f < minFreq || minIdx == -1 {
			minFreq = f
			minIdx = i
		}
	}
	return minIdx
}

func (sg *SubGraph) VertexEmbeddings(dt *Digraph, idx int) ([]*goiso.SubGraph, error) {
	embs := make([]*goiso.SubGraph, 0, dt.G.ColorFrequency(sg.V[idx].Color))
	err := dt.ColorMap.DoFind(int32(sg.V[idx].Color), func(color, dtIdx int32) error {
		sg, _ := dt.G.VertexSubGraph(int(dtIdx))
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

func (sg *SubGraph) ExtendEmbedding(dt *Digraph, cur *goiso.SubGraph, e *Edge) []*goiso.SubGraph {
	// errors.Logf("DEBUG", "extend emb %v with %v", cur.Label(), e)
	exts := make([]*goiso.SubGraph, 0, 10)
	srcs := sg.findSrcs(cur, e)
	// errors.Logf("DEBUG", "  srcs %v", srcs)
	seen := make(map[int]bool)
	for _, src := range srcs {
		for _, ke := range sg.findEdgesFromSrc(dt, cur, src, e) {
			// errors.Logf("DEBUG", "    ke %v %v", ke.Idx, ke)
			if !seen[ke.Idx] {
				seen[ke.Idx] = true
				ext, _ := cur.EdgeExtend(ke)
				exts = append(exts, ext)
			}
		}
	}
	targs := sg.findTargs(cur, e)
	// errors.Logf("DEBUG", "  targs %v", targs)
	for _, targ := range targs {
		for _, pe := range sg.findEdgesFromTarg(dt, cur, targ, e) {
			// errors.Logf("DEBUG", "    pe %v %v", pe.Idx, pe)
			if !seen[pe.Idx] {
				seen[pe.Idx] = true
				ext, _ := cur.EdgeExtend(pe)
				exts = append(exts, ext)
			}
		}
	}
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

func (sg *SubGraph) findEdgesFromSrc(dt *Digraph, cur *goiso.SubGraph, src int, e *Edge) []*goiso.Edge {
	srcDtIdx := cur.V[src].Id
	tcolor := sg.V[e.Targ].Color
	ecolor := e.Color
	edges := make([]*goiso.Edge, 0, 10)
	for _, ke := range dt.G.Kids[srcDtIdx] {
		if ke.Color != ecolor {
			continue
		} else if dt.G.V[ke.Targ].Color != tcolor {
			continue
		}
		if !cur.HasEdge(goiso.ColoredArc{ke.Arc, ke.Color}) {
			edges = append(edges, ke)
		}
	}
	return edges
}

func (sg *SubGraph) findEdgesFromTarg(dt *Digraph, cur *goiso.SubGraph, targ int, e *Edge) []*goiso.Edge {
	targDtIdx := cur.V[targ].Id
	scolor := sg.V[e.Src].Color
	ecolor := e.Color
	edges := make([]*goiso.Edge, 0, 10)
	for _, pe := range dt.G.Parents[targDtIdx] {
		if pe.Color != ecolor {
			continue
		} else if dt.G.V[pe.Src].Color != scolor {
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



