package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
)

func graph(t *testing.T) (*Graph, *goiso.Graph, *goiso.SubGraph, *SubGraph) {
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
	conf := &config.Config{
		Support: 2,
	}

	l, err := NewVegLoader(conf, MinImgSupported, 0, len(G.V), 0, len(G.E))
	if err != nil {
		t.Fatal(err)
	}
	v := l.(*VegLoader)

	// compute the starting points (we are now ready to mine)
	start, err := v.ComputeStartingPoints(G)
	if err != nil {
		t.Fatal(err)
	}
	v.G.FrequentVertices = start

	return v.G, G, sg, NewSubGraph(sg)
}

func TestEmebeddings(t *testing.T) {
	x := assert.New(t)
	dt, _, _, sg := graph(t)
	t.Log(sg)
	t.Log(sg.Adj)

	embs, err := sg.Embeddings(dt)
	x.Nil(err)
	for _, emb := range embs {
		t.Log(emb.Label())
	}
	for _, emb := range embs {
		t.Log(emb)
	}
	x.Equal(len(embs), 2, "embs should have 2 embeddings")
}
