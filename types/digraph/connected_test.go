package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"os"
	"runtime"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
)


func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if urandom, err := os.Open("/dev/urandom"); err != nil {
		panic(err)
	} else {
		seed := make([]byte, 8)
		if _, err := urandom.Read(seed); err == nil {
			rand.Seed(int64(binary.BigEndian.Uint64(seed)))
		}
		urandom.Close()
	}
}



func randomGraph(t *testing.T, V, E int, labels []string) (*Digraph, *goiso.Graph, *goiso.SubGraph, *SubGraph, *EmbListNode, *SearchNode) {
	Graph := goiso.NewGraph(10, 10)
	G := &Graph

	vidxs := make([]int, 0, V)
	vertices := make([]*goiso.Vertex, 0, V)
	for i := 0; i < V; i++ {
		v := G.AddVertex(i, labels[rand.Intn(len(labels))])
		vertices = append(vertices, v)
		vidxs = append(vidxs, v.Idx)
	}
	for i := 0; i < E; i++ {
		G.AddEdge(vertices[rand.Intn(len(vertices))], vertices[rand.Intn(len(vertices))], "")
	}

	sg, _ := G.SubGraph(vidxs, nil)

	// make config
	conf := &config.Config{
		Support: 2,
	}

	// make the *Digraph
	dt, err := NewDigraph(conf, false, MinImgSupported, 0, len(G.V), 0, len(G.E))
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		t.Fatal(err)
	}

	err = dt.Init(G)
	if err != nil {
		t.Fatal(err)
	}

	return dt, G, sg, NewSubGraph(sg), RootEmbListNode(dt), RootSearchNode(dt)
}

func TestVerifyChildrenParents(t *testing.T) {
	x := assert.New(t)
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	V := 150
	_, _, _, _, eroot, sroot := randomGraph(t, V, int(float64(V)*1.5), labels)
	dfs(t, x, eroot)
	dfs(t, x, sroot)
}

func dfs(t *testing.T, x *assert.Assertions, root Node) {
	visit(t, x, set.NewSortedSet(250), root)
}

func visit(t *testing.T, x *assert.Assertions, visited *set.SortedSet, node Node) {
	errors.Logf("DEBUG", "visiting %v", node)
	visited.Add(node.Pattern())
	checkNode(t, x, node)
	kids, err := node.Children()
	x.Nil(err)
	for _, kid := range kids {
		if !visited.Has(kid.Pattern()) {
			visit(t, x, visited, kid.(Node))
		}
	}
}

func checkNode(t *testing.T, x *assert.Assertions, node Node) {
	acount, err := node.AdjacentCount()
	x.Nil(err)
	kcount, err := node.ChildCount()
	x.Nil(err)
	kids, err := node.Children()
	x.Nil(err)
	pcount, err := node.ParentCount()
	x.Nil(err)
	parents, err := node.Parents()
	x.Nil(err)
	if kcount != len(kids) {
		x.Fail("kcount != len(kids)")
	}
	if pcount != len(parents) {
		x.Fail("count != len(parents)")
	}
	if kcount + pcount != acount {
		x.Fail("kcount + pcount != acount")
	}
	for _, kid := range kids {
		checkKid(t, x, node, kid.(Node))
	}
	for _, parent := range parents {
		checkKid(t, x, parent.(Node), node)
	}
}

func checkKid(t *testing.T, x *assert.Assertions, parent, kid Node) {
	pkids, err := parent.Children()
	x.Nil(err)
	found := false
	for _, pkid := range pkids {
		if bytes.Equal(pkid.Pattern().Label(), kid.Label()) {
			found = true
		}
	}
	if !found {
		x.Fail(errors.Errorf("parent %v kids %v did not have %v", parent, pkids, kid).Error())
	}
	kparents, err := kid.Parents()
	x.Nil(err)
	found = false
	for _, kparent := range kparents {
		if bytes.Equal(kparent.Pattern().Label(), parent.Label()) {
			found = true
		}
	}
	if !found {
		x.Fail(errors.Errorf("kid %v parents %v did not have %v", kid, kparents, parent).Error())
	}
	pkids, err = parent.CanonKids()
	x.Nil(err)
	found = false
	for _, pkid := range pkids {
		if bytes.Equal(pkid.Pattern().Label(), kid.Label()) {
			found = true
		}
	}
	if found {
		// kid is a canon kid
		// kid should have no other canon parents
		kparents, err := kid.Parents()
		x.Nil(err)
		found = false
		for _, kparent := range kparents {
			if bytes.Equal(kparent.Pattern().Label(), parent.Label()) {
				continue
			}
			kparent_ckids, err := kparent.CanonKids()
			x.Nil(err)
			for _, kparent_ckid := range kparent_ckids {
				if bytes.Equal(kparent_ckid.Pattern().Label(), kid.Label()) {
					x.Fail(errors.Errorf("kid %v had multiple canon parents %v %v", kid, parent, kparent).Error())
				}
			}
		}
	}
}
