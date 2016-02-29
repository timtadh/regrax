package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"github.com/timtadh/data-structures/errors"
)

import (
)

func TestSearchChildren(t *testing.T) {
	x := assert.New(t)
	_, _, _, _, _, n := graph(t)
	x.NotNil(n)
	t.Log(n)
	kids, err := n.Children()
	x.Nil(err)
	var next *SearchNode = nil
	for _, k := range kids {
		t.Log(k)
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
	return
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


func TestSearchCount(t *testing.T) {
	x := assert.New(t)
	_, _, _, _, _, n := graph(t)
	x.NotNil(n)
	count, err := n.ChildCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 kids")
	count, err = n.ParentCount()
	x.Nil(err)
	x.Equal(count, 0, "should have 0 parents")
	count, err = n.ChildCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 kids")
	kids, err := n.Children()
	x.Nil(err)
	var next *SearchNode = nil
	for _, k := range kids {
		kid := k.(*SearchNode)
		switch kid.String() {
		case "<SearchNode 0:1(0:red)>":
		case "<SearchNode 0:1(0:black)>":
			next = kid
		default: x.Fail(errors.Errorf("unexpected kid %v", kid).Error())
		}
	}
	if next == nil {
		x.Fail("did not find the black node")
	}
	cur := next
	next = nil
	count, err = cur.ChildCount()
	x.Nil(err)
	x.Equal(count, 1, "should have 1 kids")
	count, err = cur.ParentCount()
	x.Nil(err)
	x.Equal(count, 1, "should have 1 parents")
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	cur = kids[0].(*SearchNode)
	count, err = cur.ChildCount()
	x.Nil(err)
	x.Equal(count, 3, "should have 3 kids")
	count, err = cur.ParentCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 parents")
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	x.Equal(len(kids), 3, "should have 3 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*SearchNode)
		switch kid.String() {
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->1:][0->2:]>":
			next = kid
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->1:][2->1:]>":
		case "<SearchNode 2:3(0:black)(1:red)(2:red)[0->2:][2->1:]>":
		default: x.Fail("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	count, err = cur.ChildCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 kids")
	count, err = cur.ParentCount()
	x.Nil(err)
	x.Equal(count, 1, "should have 1 parents")
	kids, err = cur.Children()
	if err != nil {
		t.Log(err)
		x.Nil(err)
	}
	x.Equal(len(kids), 2, "should have 2 kids got %v", kids)
	for _, k := range kids {
		kid := k.(*SearchNode)
		switch kid.String() {
		case "<SearchNode 3:4(0:black)(1:red)(2:red)(3:red)[0->1:][0->2:][3->2:]>":
			next = kid
		case "<SearchNode 3:4(0:black)(1:red)(2:red)(3:red)[0->1:][0->3:][3->2:]>":
		default: x.Fail("unexpected kid %v", kid)
		}
	}
	cur = next
	next = nil
	count, err = cur.ChildCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 kids")
	count, err = cur.ParentCount()
	x.Nil(err)
	x.Equal(count, 2, "should have 2 parents")
	/// stopping this exercise here.
}

