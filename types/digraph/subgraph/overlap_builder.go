package subgraph

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
)

type OverlapBuilder struct {
	*Builder
	Adj [][]int
	Ids []*set.SortedSet
}

type FillableOverlapBuilder struct {
	*OverlapBuilder
}

func BuildOverlap(V, E int) *OverlapBuilder {
	return &OverlapBuilder{
		Builder: &Builder{
			V: make([]Vertex, 0, V),
			E: make([]Edge, 0, E),
		},
		Adj: make([][]int, 0, V),
		Ids: make([]*set.SortedSet, 0, V),
	}
}

// not implemented
func (b *OverlapBuilder) From(sg *SubGraph) *OverlapBuilder {
	panic("not-implemented")
}

// not implemented
func (b *OverlapBuilder) FromVertex(color int) *OverlapBuilder {
	panic("not-implemented")
}

func (b *OverlapBuilder) Fillable() *FillableOverlapBuilder {
	if len(b.V) != 0 || len(b.E) != 0 || len(b.Ids) != 0 {
		panic("embedding builder must be empty to use Fillable")
	}
	b.V = b.V[:cap(b.V)]
	b.Ids = b.Ids[:cap(b.Ids)]
	b.Adj = b.Adj[:cap(b.Adj)]
	for i := range b.V {
		b.V[i].Idx = -1
	}
	for i := range b.Ids {
		b.Ids[i] = set.NewSortedSet(1)
	}
	for i := range b.Adj {
		b.Adj[i] = make([]int, 0, 1)
	}
	return &FillableOverlapBuilder{b}
}

func (b *OverlapBuilder) Copy() *OverlapBuilder {
	adj := make([][]int, len(b.Adj))
	for i := range adj {
		a := make([]int, len(b.Adj[i]))
		copy(a, b.Adj[i])
		adj[i] = a
	}
	ids := make([]*set.SortedSet, len(b.Ids))
	for i := range ids {
		ids[i] = b.Ids[i].Copy()
	}
	return &OverlapBuilder{
		Builder: b.Builder.Copy(),
		Adj:     adj,
		Ids:     ids,
	}
}

func (b *OverlapBuilder) Ctx(do func(*OverlapBuilder)) *OverlapBuilder {
	do(b)
	return b
}

func (b *OverlapBuilder) Do(do func(*OverlapBuilder) error) (*OverlapBuilder, error) {
	err := do(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *OverlapBuilder) AddVertex(color int, ids *set.SortedSet) *Vertex {
	b.V = append(b.V, Vertex{
		Idx:   len(b.V),
		Color: color,
	})
	b.Ids = append(b.Ids, ids)
	return &b.V[len(b.V)-1]
}

func (b *OverlapBuilder) AddEdge(src, targ *Vertex, color int) *Edge {
	e := b.Builder.AddEdge(src, targ, color)
	eidx := len(b.E) - 1
	b.Adj[e.Src] = append(b.Adj[e.Src], eidx)
	if e.Src != e.Targ {
		b.Adj[e.Targ] = append(b.Adj[e.Targ], eidx)
	}
	return e
}

func (b *OverlapBuilder) RemoveEdge(edgeIdx int) error {
	return errors.Errorf("not-implemented")
}

func (b *OverlapBuilder) Extend(e *Extension) (newe *Edge, newv *Vertex, err error) {
	return nil, nil, errors.Errorf("not-implemented")
}

func (b *OverlapBuilder) Build() *Overlap {
	vord, eord := b.canonicalPermutation()
	sg := b.build(vord, eord)
	ids := make([]*set.SortedSet, len(sg.V))
	for i, p := range vord {
		ids[p] = b.Ids[i]
	}
	return &Overlap{SG: sg, Ids: ids}
}

func (b *FillableOverlapBuilder) SetVertex(idx, color int, ids *set.SortedSet) {
	b.V[idx].Idx = idx
	b.V[idx].Color = color
	b.Ids[idx] = ids
}

func (b *FillableOverlapBuilder) Copy() *FillableOverlapBuilder {
	return &FillableOverlapBuilder{b.OverlapBuilder.Copy()}
}

func (b *FillableOverlapBuilder) Ctx(do func(*FillableOverlapBuilder)) *FillableOverlapBuilder {
	do(b)
	return b
}

func (b *FillableOverlapBuilder) String() string {
	return fmt.Sprintf("<FOB %v %v %v>", b.Builder, b.Adj, b.Ids)
}
