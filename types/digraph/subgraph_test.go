package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
)

func graph(t *testing.T) (*Digraph, *goiso.Graph, *goiso.SubGraph, *SubGraph, *EmbListNode, *SearchNode) {
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

	// make the *Digraph
	dt, err := NewDigraph(conf, false, MinImgSupported, 0, len(G.V), 0, len(G.E))
	if err != nil {
		t.Fatal(err)
	}

	err = dt.Init(G)
	if err != nil {
		t.Fatal(err)
	}

	return dt, G, sg, NewSubGraph(sg), RootEmbListNode(dt), RootSearchNode(dt)
}

func TestEmebeddings(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x,x)
	dt, _, _, sg, _, _ := graph(t)
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
