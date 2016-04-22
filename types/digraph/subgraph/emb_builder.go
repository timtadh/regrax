package subgraph

import (
	"github.com/timtadh/data-structures/errors"
)


type EmbeddingBuilder struct {
	*Builder
	Ids []int // Idx vertices in the graph
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

func (b *EmbeddingBuilder) Copy() *EmbeddingBuilder {
	ids := make([]int, len(b.Ids))
	copy(ids, b.Ids)
	return &EmbeddingBuilder{
		Builder: b.Builder.Copy(),
		Ids: ids,
	}
}

func (b *EmbeddingBuilder) Mutation(do func(*EmbeddingBuilder)) *EmbeddingBuilder {
	nb := b.Copy()
	do(nb)
	return nb
}

func (b *EmbeddingBuilder) From(emb *Embedding) {
	if len(b.V) != 0 || len(b.E) != 0 || len(b.Ids) != 0 {
		panic("embedding builder must be empty to use From")
	}
	for i := range emb.V {
		b.AddVertex(emb.V[i].Color, emb.Ids[i])
	}
	for i := range emb.E {
		b.AddEdge(&emb.V[emb.E[i].Src], &emb.V[emb.E[i].Targ], emb.E[i].Color)
	}
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
	vord, eord := b.canonicalPermutation()
	sg := b.build(vord, eord)
	ids := make([]int, len(sg.V))
	for i, p := range vord {
		ids[p] = b.Ids[i]
	}
	return &Embedding{SubGraph: sg, Ids: ids}
}



