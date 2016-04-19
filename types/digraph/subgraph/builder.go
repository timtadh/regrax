package subgraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso/bliss"
)

type Builder struct {
	V Vertices
	E Edges
}

func BuildNew() *Builder {
	return &Builder{
		V: make([]Vertex, 0, 10),
		E: make([]Edge, 0, 10),
	}
}

func BuildFrom(sg *SubGraph) *Builder {
	V := make([]Vertex, len(sg.V))
	E := make([]Edge, len(sg.E))
	copy(V, sg.V)
	copy(E, sg.E)
	return &Builder{
		V: V,
		E: E,
	}
}

func BuildFromVertex(color int) *Builder {
	b := BuildNew()
	b.AddVertex(color)
	return b
}

func (b *Builder) Copy() *Builder {
	V := make([]Vertex, len(b.V))
	E := make([]Edge, len(b.E))
	copy(V, b.V)
	copy(E, b.E)
	return &Builder{
		V: V,
		E: E,
	}
}

func (b *Builder) Mutation(do func(*Builder)) *Builder {
	nb := b.Copy()
	do(nb)
	return nb
}

func (b *Builder) AddVertex(color int) *Vertex {
	b.V = append(b.V, Vertex{
		Idx:   len(b.V),
		Color: color,
	})
	return &b.V[len(b.V)-1]
}

func (b *Builder) AddEdge(src, targ *Vertex, color int) *Edge {
	b.E = append(b.E, Edge{
		Src:   src.Idx,
		Targ:  targ.Idx,
		Color: color,
	})
	return &b.E[len(b.E)-1]
}

func (b *Builder) RemoveEdge(edgeIdx int) error {
	edge := &b.E[edgeIdx]
	rmSrc := true
	rmTarg := true
	for i := range b.E {
		e := &b.E[i]
		if e == edge {
			continue
		}
		if edge.Src == e.Src || edge.Src == e.Targ {
			// a kid edge
			rmSrc = false
		}
		if edge.Targ == e.Src || edge.Targ == e.Targ {
			// a parent edge
			rmTarg = false
		}
	}
	if rmSrc && rmTarg {
		return errors.Errorf("would have removed both source and target %v %v", rmSrc, rmTarg)
	}
	rmV := rmSrc || rmTarg
	var rmVidx int
	if rmSrc {
		rmVidx = edge.Src
	}
	if rmTarg {
		rmVidx = edge.Targ
	}
	adjustIdx := func(idx int) int {
		if rmV && idx > rmVidx {
			return idx - 1
		}
		return idx
	}
	V := make([]Vertex, 0, len(b.V))
	for idx := range b.V {
		if rmV && rmVidx == idx {
			continue
		}
		V = append(V, Vertex{Idx:adjustIdx(idx), Color:b.V[idx].Color})
	}
	E := make([]Edge, 0, len(b.E)-1)
	for idx := range b.E {
		if idx == edgeIdx {
			continue
		}
		E = append(E, Edge{
			Src:adjustIdx(b.E[idx].Src),
			Targ:adjustIdx(b.E[idx].Targ),
			Color:b.E[idx].Color,
		})
	}
	b.V = V
	b.E = E
	return nil
}

func (b *Builder) Extend(e *Extension) (newe *Edge, newv *Vertex, err error) {
	if e.Source.Idx > len(b.V) {
		return nil, nil, errors.Errorf("Source.Idx %v outside of |V| %v", e.Source.Idx, len(b.V))
	} else if e.Target.Idx > len(b.V) {
		return nil, nil, errors.Errorf("Target.Idx %v outside of |V| %v", e.Target.Idx, len(b.V))
	} else if e.Source.Idx == len(b.V) && e.Target.Idx == len(b.V) {
		return nil, nil, errors.Errorf("Only one new vertice allowed (Extension would create a disconnnected graph)")
	}
	var src *Vertex = &e.Source
	var targ *Vertex = &e.Target
	if e.Source.Idx == len(b.V) {
		src = b.AddVertex(e.Source.Color)
		newv = src
	} else if e.Target.Idx == len(b.V) {
		targ = b.AddVertex(e.Target.Color)
		newv = targ
	}
	newe = b.AddEdge(src, targ, e.Color)
	return newe, newv, nil
}

func (b *Builder) Kids() [][]*Edge {
	kids := make([][]*Edge, 0, len(b.V))
	for _ = range b.V {
		kids = append(kids, make([]*Edge, 0, 5))
	}
	for i := range b.E {
		e := &b.E[i]
		kids[e.Src] = append(kids[e.Src], e)
	}
	return kids
}

func (b *Builder) Parents() [][]*Edge {
	parents := make([][]*Edge, 0, len(b.V))
	for _ = range b.V {
		parents = append(parents, make([]*Edge, 0, 5))
	}
	for i := range b.E {
		e := &b.E[i]
		parents[e.Targ] = append(parents[e.Targ], e)
	}
	return parents
}

func (b *Builder) Connected() bool {
	kids := b.Kids()
	parents := b.Parents()
	pop := func(stack []int) (int, []int) {
		idx := stack[len(stack)-1]
		stack = stack[0 : len(stack)-1]
		return idx, stack
	}
	visit := func(idx int, stack []int, processed map[int]bool) []int {
		processed[idx] = true
		for _, kid := range kids[idx] {
			if _, has := processed[kid.Targ]; !has {
				stack = append(stack, kid.Targ)
			}
		}
		for _, parent := range parents[idx] {
			if _, has := processed[parent.Src]; !has {
				stack = append(stack, parent.Src)
			}
		}
		return stack
	}
	processed := make(map[int]bool, len(b.V))
	stack := make([]int, 0, len(b.V))
	stack = append(stack, 0)
	for len(stack) > 0 {
		var v int
		v, stack = pop(stack)
		stack = visit(v, stack, processed)
	}
	return len(processed) == len(b.V)
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
