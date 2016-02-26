package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	// "github.com/timtadh/goiso"
)

import (
	// "github.com/timtadh/sfp/config"
)

func TestChildren(t *testing.T) {
	x := assert.New(t)
	dt, _, _, _ := graph(t)
	n := dt.Empty()
	x.NotNil(n)
	kids, err := n.Children()
	x.Nil(err)
	var next *SearchNode = nil
	for _, k := range kids {
		kid := k.(*SearchNode)
		embs, err := kid.Embeddings()
		x.Nil(err)
		switch kid.String() {
		case "<SearchNode 0:1(0:red)>":
			x.Equal(len(embs), 4, "4 embeddings")
		case "<SearchNode 0:1(0:black)>":
			x.Equal(len(embs), 2, "2 embeddings")
			next = kid
		default: x.Fail("unexpected kid %v", kid)
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
		x.Nil(err)
	}
	x.Equal(len(kids), 1, "should have 1 kids 1:2(0:black)(1:red)[0->1:] got %v", kids)
	cur = kids[0].(*SearchNode)
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	x.Equal(len(kids), 3, "should have 3 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*SearchNode)
		embs, err := kid.Embeddings()
		x.Nil(err)
		switch kid.String() {
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->1:][0->2:]>":
			x.Equal(len(embs), 4, "4 embeddings")
			next = kid
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->1:][2->1:]>":
			x.Equal(len(embs), 2, "2 embeddings")
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->2:][2->1:]>":
			x.Equal(len(embs), 2, "2 embeddings")
		default: x.Fail("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*SearchNode)
		embs, err := kid.Embeddings()
		x.Nil(err)
		switch kid.String() {
		case "<SearchNode 3:4(0:black)(1:red)(2:red)(3:red)[0->1:][0->2:][3->2:]>":
			x.Equal(len(embs), 2, "2 embeddings")
			next = kid
		case "<SearchNode 3:4(0:black)(1:red)(2:red)(3:red)[0->1:][0->3:][3->2:]>":
			x.Equal(len(embs), 2, "2 embeddings")
		default: x.Fail("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	/// stopping this exercise here.
}

