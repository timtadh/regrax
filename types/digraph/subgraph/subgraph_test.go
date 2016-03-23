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
	"github.com/timtadh/sfp/config"
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
	conf := &config.Config{}
	ColorMap, err := conf.IntIntMultiMap("color-map")
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

	return G, sg, NewSubGraph(sg), ColorMap, ext.NewExtender(runtime.NumCPU())
}

func TestEmebeddings(t *testing.T) {
	x := assert.New(t)
	t.Logf("%T %v", x,x)
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
