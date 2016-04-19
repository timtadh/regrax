package subgraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"runtime"
)

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/types/digraph/ext"
)

func graph(t *testing.T) (*goiso.Graph, *goiso.SubGraph, *SubGraph, int_int.MultiMap, *ext.Extender) {
	Graph := goiso.NewGraph(10, 10)
	G := &Graph
	n1 := G.AddVertex(1, "black")
	n2 := G.AddVertex(2, "black")
	n3 := G.AddVertex(3, "red")
	n4 := G.AddVertex(4, "red")
	n5 := G.AddVertex(5, "red")
	n6 := G.AddVertex(6, "red")
	G.AddEdge(n1, n3, "")
	G.AddEdge(n1, n4, "")
	G.AddEdge(n2, n5, "")
	G.AddEdge(n2, n6, "")
	G.AddEdge(n5, n3, "")
	G.AddEdge(n4, n6, "")
	sg, _ := G.SubGraph([]int{n1.Idx, n2.Idx, n3.Idx, n4.Idx, n5.Idx, n6.Idx}, nil)

	// make config
	ColorMap, err := int_int.AnonBpTree()
	if err != nil {
		t.Fatal(err)
	}

	for i := range G.V {
		u := &G.V[i]
		err := ColorMap.Add(int32(u.Color), int32(u.Idx))
		if err != nil {
			t.Fatal(err)
		}
	}

	return G, sg, FromEmbedding(sg), ColorMap, ext.NewExtender(runtime.NumCPU())
}

func TestEmbeddings(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x, x)
	G, _, sg, colors, extender := graph(t)
	t.Log(sg)
	t.Log(sg.Adj)

	embs, err := sg.Embeddings(G, colors, extender)
	x.Nil(err)
	for _, emb := range embs {
		t.Log(emb.Label())
	}
	for _, emb := range embs {
		t.Log(emb)
	}
	x.Equal(len(embs), 2, "embs should have 2 embeddings")
}

func TestNewBuilder(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x, x)
	G, _, expected, colors, extender := graph(t)
	b := BuildNew()
	n1 := b.AddVertex(0)
	n2 := b.AddVertex(0)
	n3 := b.AddVertex(1)
	n4 := b.AddVertex(1)
	n5 := b.AddVertex(1)
	n6 := b.AddVertex(1)
	b.AddEdge(n1, n3, 2)
	b.AddEdge(n1, n4, 2)
	b.AddEdge(n2, n5, 2)
	b.AddEdge(n5, n3, 2)
	b.AddEdge(n4, n6, 2)
	b.AddEdge(n2, n6, 2)
	sg := b.Build()
	t.Log(sg)
	t.Log(sg.Adj)
	t.Log(expected)
	t.Log(expected.Adj)
	x.Equal(sg.String(), expected.String())

	embs, err := sg.Embeddings(G, colors, extender)
	x.Nil(err)
	for _, emb := range embs {
		t.Log(emb.Label())
	}
	for _, emb := range embs {
		t.Log(emb)
	}
	x.Equal(len(embs), 2, "embs should have 2 embeddings")
}

func TestFromBuilder(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x, x)
	G, _, expected, colors, extender := graph(t)
	b := BuildNew()
	n1 := b.AddVertex(0)
	n2 := b.AddVertex(0)
	n3 := b.AddVertex(1)
	n4 := b.AddVertex(1)
	n5 := b.AddVertex(1)
	n6 := b.AddVertex(1)
	b.AddEdge(n1, n3, 2)
	b.AddEdge(n1, n4, 2)
	b.AddEdge(n2, n5, 2)
	b.AddEdge(n2, n6, 2)
	b.AddEdge(n4, n6, 2)
	sg1 := b.Build()
	t.Log(sg1)
	t.Log(sg1.Adj)
	b2 := BuildFrom(sg1)
	sg := b2.Mutation(func(b *Builder) { b.AddEdge(&b.V[3], &b.V[5], 2) }).Build()
	t.Log(b2.Build())
	t.Log(sg)
	t.Log(expected)
	t.Log(sg.Adj)
	t.Log(expected.Adj)
	x.Equal(sg.String(), expected.String())

	embs, err := sg.Embeddings(G, colors, extender)
	x.Nil(err)
	for _, emb := range embs {
		t.Log(emb.Label())
	}
	for _, emb := range embs {
		t.Log(emb)
	}
	x.Equal(len(embs), 2, "embs should have 2 embeddings")
}

func TestFromExtension(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x, x)
	G, _, expected, colors, extender := graph(t)
	b := BuildNew()
	n1 := b.AddVertex(0)
	n2 := b.AddVertex(0)
	n3 := b.AddVertex(1)
	n4 := b.AddVertex(1)
	n5 := b.AddVertex(1)
	b.AddEdge(n1, n3, 2)
	b.AddEdge(n1, n4, 2)
	b.AddEdge(n2, n5, 2)
	_, n6, _ := b.Extend(NewExt(*n2, Vertex{Idx:5, Color:1}, 2))
	b.Extend(NewExt(*n4, *n6, 2))
	b.Extend(NewExt(*n5, *n3, 2))
	sg := b.Build()
	t.Log(sg)
	t.Log(expected)
	t.Log(sg.Adj)
	t.Log(expected.Adj)
	x.Equal(sg.String(), expected.String())

	embs, err := sg.Embeddings(G, colors, extender)
	x.Nil(err)
	for _, emb := range embs {
		t.Log(emb.Label())
	}
	for _, emb := range embs {
		t.Log(emb)
	}
	x.Equal(len(embs), 2, "embs should have 2 embeddings")
}

func TestBuilderRemoveEdge(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x, x)
	_, _, expected, _, _ := graph(t)
	b := BuildNew()
	n1 := b.AddVertex(0)
	n2 := b.AddVertex(0)
	n3 := b.AddVertex(1)
	n4 := b.AddVertex(1)
	n5 := b.AddVertex(1)
	n6 := b.AddVertex(1)
	b.AddEdge(n1, n3, 2)
	b.AddEdge(n1, n4, 2)
	b.AddEdge(n2, n5, 2)
	b.AddEdge(n5, n3, 2)
	b.AddEdge(n4, n6, 2)
	b.AddEdge(n2, n6, 2)
	b.AddEdge(n3, n6, 2)
	wrong := b.Build()
	t.Log(wrong)
	t.Log(expected)
	x.NotEqual(wrong.String(), expected.String())
	b = BuildFrom(wrong)
	err := b.RemoveEdge(6)
	if err != nil {
		t.Fatal(err)
	}
	right := b.Build()
	t.Log(right)
	t.Log(expected)
	x.Equal(right.String(), expected.String())
}

func TestBuilderConnected(t *testing.T) {
	x := assert.New(t)
	b := BuildNew()
	n1 := b.AddVertex(0)
	n2 := b.AddVertex(0)
	n3 := b.AddVertex(1)
	n4 := b.AddVertex(1)
	n5 := b.AddVertex(1)
	n6 := b.AddVertex(1)
	b.AddEdge(n1, n3, 2)
	b.AddEdge(n1, n4, 2)
	b.AddEdge(n2, n5, 2)
	b.AddEdge(n5, n3, 2)
	b.AddEdge(n4, n6, 2)
	b.AddEdge(n2, n6, 2)
	x.True(b.Connected())
	_ = b.AddVertex(2)
	x.False(b.Connected())
}

