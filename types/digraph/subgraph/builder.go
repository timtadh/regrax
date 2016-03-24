package subgraph

import (
	"github.com/timtadh/goiso/bliss"
)


type Builder struct {
	V Vertices
	E Edges
}

func New() *Builder {
	return &Builder{
		V: make([]Vertex, 0, 10),
		E: make([]Edge, 0, 10),
	}
}

func From(sg *SubGraph) *Builder {
	V := make([]Vertex, len(sg.V))
	E := make([]Edge, len(sg.E))
	copy(V, sg.V)
	copy(E, sg.E)
	return &Builder{
		V: V,
		E: E,
	}
}

func (b *Builder) AddVertex(color int) *Vertex {
	b.V = append(b.V, Vertex{
		Idx: len(b.V),
		Color: color,
	})
	return &b.V[len(b.V)-1]
}

func (b *Builder) AddEdge(src, targ *Vertex, color int) *Edge {
	b.E = append(b.E, Edge{
		Src: src.Idx,
		Targ: targ.Idx,
		Color: color,
	})
	return &b.E[len(b.E)-1]
}

func (b *Builder) Build() *SubGraph {
	pat := &SubGraph{
		V:   make([]Vertex, len(b.V)),
		E:   make([]Edge, len(b.E)),
		Adj: make([][]int, len(b.V)),
	}
	bMap := bliss.NewMap(len(b.V), len(b.E), b.V.Iterate(), b.E.Iterate())
	vord, eord, _ := bMap.CanonicalPermutation()
	for i, j := range vord {
		pat.V[j].Idx = b.V[i].Idx
		pat.V[j].Color = b.V[i].Color
		pat.Adj[j] = make([]int, 0, 5)
	}
	for i, j := range eord {
		pat.E[j].Src = vord[b.E[i].Src]
		pat.E[j].Targ = vord[b.E[i].Targ]
		pat.E[j].Color = b.E[i].Color
		pat.Adj[pat.E[j].Src] = append(pat.Adj[pat.E[j].Src], j)
		pat.Adj[pat.E[j].Targ] = append(pat.Adj[pat.E[j].Targ], j)
	}
	return pat
}

