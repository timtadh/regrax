package subgraph

import (
	"github.com/timtadh/data-structures/errors"
)

type EmbeddingBuilder struct {
	*Builder
	Ids []int // Idx vertices in the graph
}

type FillableEmbeddingBuilder struct {
	*EmbeddingBuilder
}

func BuildEmbedding(V, E int) *EmbeddingBuilder {
	return &EmbeddingBuilder{
		Builder: &Builder{
			V: make([]Vertex, 0, V),
			E: make([]Edge, 0, E),
		},
		Ids: make([]int, 0, V),
	}
}

func (b *EmbeddingBuilder) Fillable() *FillableEmbeddingBuilder {
	if len(b.V) != 0 || len(b.E) != 0 || len(b.Ids) != 0 {
		panic("embedding builder must be empty to use Fillable")
	}
	b.V = b.V[:cap(b.V)]
	b.Ids = b.Ids[:cap(b.Ids)]
	for i := range b.V {
		b.V[i].Idx = -1
		b.Ids[i] = -1
	}
	return &FillableEmbeddingBuilder{b}
}

// Mutates the current builder and returns it
func (b *EmbeddingBuilder) From(emb *Embedding) *EmbeddingBuilder {
	if len(b.V) != 0 || len(b.E) != 0 || len(b.Ids) != 0 {
		panic("embedding builder must be empty to use From")
	}
	for i := range emb.SG.V {
		b.AddVertex(emb.SG.V[i].Color, emb.Ids[i])
	}
	for i := range emb.SG.E {
		b.AddEdge(&emb.SG.V[emb.SG.E[i].Src], &emb.SG.V[emb.SG.E[i].Targ], emb.SG.E[i].Color)
	}
	return b
}

func (b *EmbeddingBuilder) FromVertex(color, id int) *EmbeddingBuilder {
	b.AddVertex(color, id)
	return b
}

func (b *EmbeddingBuilder) Copy() *EmbeddingBuilder {
	ids := make([]int, len(b.Ids))
	copy(ids, b.Ids)
	return &EmbeddingBuilder{
		Builder: b.Builder.Copy(),
		Ids:     ids,
	}
}

func (b *EmbeddingBuilder) Ctx(do func(*EmbeddingBuilder)) *EmbeddingBuilder {
	do(b)
	return b
}

func (b *EmbeddingBuilder) Do(do func(*EmbeddingBuilder) error) (*EmbeddingBuilder, error) {
	err := do(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *EmbeddingBuilder) HasId(id int) bool {
	for _, x := range b.Ids {
		if id == x {
			return true
		}
	}
	return false
}

func (b *EmbeddingBuilder) AddVertex(color, id int) *Vertex {
	b.V = append(b.V, Vertex{
		Idx:   len(b.V),
		Color: color,
	})
	b.Ids = append(b.Ids, id)
	return &b.V[len(b.V)-1]
}

func (b *EmbeddingBuilder) RemoveEdge(edgeIdx int) error {
	dropVertex, vertexIdx, err := b.droppedVertexOnEdgeRm(edgeIdx)
	if err != nil {
		return err
	}
	b.V, b.E = b.removeVertexAndEdge(dropVertex, vertexIdx, edgeIdx)
	if dropVertex {
		ids := make([]int, 0, len(b.V))
		for idx, id := range b.Ids {
			if idx == vertexIdx {
				continue
			}
			ids = append(ids, id)
		}
		b.Ids = ids
	}
	return nil
}

func (b *EmbeddingBuilder) Extend(e *Extension) (newe *Edge, newv *Vertex, err error) {
	return nil, nil, errors.Errorf("not-implemented")
}

func (b *EmbeddingBuilder) Build() *Embedding {
	vord, eord := b.CanonicalPermutation()
	sg := b.BuildFromPermutation(vord, eord)
	ids := make([]int, len(sg.V))
	for i, p := range vord {
		ids[p] = b.Ids[i]
	}
	return &Embedding{SG: sg, Ids: ids}
}

func (b *FillableEmbeddingBuilder) SetVertex(idx, color, id int) {
	b.V[idx].Idx = idx
	b.V[idx].Color = color
	b.Ids[idx] = id
}

func (b *FillableEmbeddingBuilder) Copy() *FillableEmbeddingBuilder {
	return &FillableEmbeddingBuilder{b.EmbeddingBuilder.Copy()}
}

func (b *FillableEmbeddingBuilder) Ctx(do func(*FillableEmbeddingBuilder)) *FillableEmbeddingBuilder {
	do(b)
	return b
}
