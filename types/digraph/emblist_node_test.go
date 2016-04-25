package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

func graph(t *testing.T) (*Digraph, *goiso.Graph, *goiso.SubGraph, *subgraph.SubGraph, *EmbListNode) {
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
	dt, err := NewDigraph(conf, MinImgSupported, 0, len(G.V), 0, len(G.E))
	if err != nil {
		t.Fatal(err)
	}

	err = dt.Init(G)
	if err != nil {
		t.Fatal(err)
	}

	return dt, G, sg, subgraph.FromEmbedding(sg), RootEmbListNode(dt)
}

func TestEmbChildren(t *testing.T) {
	x := assert.New(t)
	_, _, _, _, n := graph(t)
	x.NotNil(n)
	kids, err := n.Children()
	if err != nil {
		t.Fatal(err)
	}
	var next *EmbListNode = nil
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {0:1}(red)>":
			x.Equal(len(kid.embeddings), 4, "4 embeddings")
		case "<EmbListNode {0:1}(black)>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		default:
			x.Fail(errors.Errorf("unexpected kid %v", kid).Error())
		}
	}
	if next == nil {
		x.Fail("did not find the black node")
	}
	cur := next
	next = nil
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 1, "should have 1 kids {1:2}(black)(red)[0->1:] got %v", kids)
	cur = kids[0].(*EmbListNode)
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 3, "should have 3 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {2:3}(black)(red)(red)[0->1:][0->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		case "<EmbListNode {2:3}(black)(red)(red)[0->1:][2->1:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		case "<EmbListNode {2:3}(black)(red)(red)[0->2:][2->1:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		default:
			t.Fatalf("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {3:4}(black)(red)(red)(red)[0->1:][0->2:][3->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		case "<EmbListNode {3:4}(black)(red)(red)(red)[0->1:][0->3:][3->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		default:
			t.Fatalf("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	/// stopping this exercise here.
}

func TestEmbCount(t *testing.T) {
	x := assert.New(t)
	_, _, _, _, n := graph(t)
	x.NotNil(n)
	count, err := n.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 kids")
	count, err = n.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 0, "should have 0 parents")
	count, err = n.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 kids")
	kids, err := n.Children()
	if err != nil {
		t.Fatal(err)
	}
	var next *EmbListNode = nil
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {0:1}(red)>":
			x.Equal(len(kid.embeddings), 4, "4 embeddings")
		case "<EmbListNode {0:1}(black)>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		default:
			x.Fail(errors.Errorf("unexpected kid %v", kid).Error())
		}
	}
	if next == nil {
		x.Fail("did not find the black node")
	}
	cur := next
	next = nil
	count, err = cur.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 1, "should have 1 kids")
	count, err = cur.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 1, "should have 1 parents")
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	cur = kids[0].(*EmbListNode)
	count, err = cur.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 3, "should have 3 kids")
	count, err = cur.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 parents")
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 3, "should have 3 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {2:3}(black)(red)(red)[0->1:][0->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		case "<EmbListNode {2:3}(black)(red)(red)[0->1:][2->1:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		case "<EmbListNode {2:3}(black)(red)(red)[0->2:][2->1:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		default:
			t.Fatalf("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	count, err = cur.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 kids")
	count, err = cur.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("cur %v should have 1 parents", cur)
		t.Logf("cur parents")
		parents, err := cur.Parents()
		if err != nil {
			t.Error(err)
		}
		for _, p := range parents {
			t.Logf("parent %v", p)
		}
	}
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*EmbListNode)
		switch kid.String() {
		case "<EmbListNode {3:4}(black)(red)(red)(red)[0->1:][0->2:][3->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
			next = kid
		case "<EmbListNode {3:4}(black)(red)(red)(red)[0->1:][0->3:][3->2:]>":
			x.Equal(len(kid.embeddings), 2, "2 embeddings")
		default:
			t.Fatalf("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	count, err = cur.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 kids")
	count, err = cur.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	x.Equal(count, 2, "should have 2 parents")
	/// stopping this exercise here.
}
