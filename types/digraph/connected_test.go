package digraph

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"fmt"
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
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/types/digraph/subgraph"
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

func randomGraph(t testing.TB, V, E int, vlabels, elabels []string) (*Digraph, *goiso.Graph, *goiso.SubGraph, *subgraph.SubGraph, *EmbListNode) {
	Graph := goiso.NewGraph(10, 10)
	G := &Graph

	vidxs := make([]int, 0, V)
	vertices := make([]*goiso.Vertex, 0, V)
	for i := 0; i < V; i++ {
		v := G.AddVertex(i, vlabels[rand.Intn(len(vlabels))])
		vertices = append(vertices, v)
		vidxs = append(vidxs, v.Idx)
	}
	for i := 0; i < E; i++ {
		src := rand.Intn(len(vertices))
		targ := rand.Intn(len(vertices))
		elabel := rand.Intn(len(elabels))
		if !G.HasEdge(vertices[src], vertices[targ], elabels[elabel]) {
			G.AddEdge(vertices[src], vertices[targ], elabels[elabel])
		}
	}

	sg, _ := G.SubGraph(vidxs, nil)

	// make config
	conf := &config.Config{
		Support: 2,
		Parallelism: -1,
	}

	// make the *Digraph
	dt, err := NewDigraph(conf, &Config{
		MinEdges: 0,
		MaxEdges: len(G.E),
		MinVertices: 0,
		MaxVertices: len(G.V),
		Mode: Automorphs | Caching | ExtensionPruning | EmbeddingPruning | ExtFromFreqEdges,
		EmbSearchStartPoint: subgraph.RandomStart,
	})
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		t.Fatal(err)
	}

	err = dt.Init(G)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("/tmp/random-graph.dot")
	if err == nil {
		fmt.Fprintf(f, "%v", G)
		f.Close()
	}

	return dt, G, sg, subgraph.FromEmbedding(sg), RootEmbListNode(dt)
}

func BenchmarkEmbList(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := assert.New(b)
		vlabels := []string{"a", "b", "c", "d", "e", "f"}
		elabels := []string{"g", "h", "i"}
		V := 100
		_, _, _, _, eroot := randomGraph(
			b, V, int(float64(V)*2.25), vlabels, elabels)
		b.StartTimer()
		dfs(b, x, eroot)
		b.StopTimer()
	}
}

func TestVerifyEmbList(t *testing.T) {
	x := assert.New(t)
	vlabels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	elabels := []string{"j", "k"}
	V := 350
	_, _, _, _, eroot := randomGraph(t, V, int(float64(V)*1.5), vlabels, elabels)
	dfs(t, x, eroot)
}

func dfs(t testing.TB, x *assert.Assertions, root Node) {
	visit(t, x, set.NewSortedSet(250), root)
}

func visit(t testing.TB, x *assert.Assertions, visited *set.SortedSet, node Node) {
	errors.Logf("DEBUG", "visiting %v", node)
	if visited.Has(node.Pattern()) {
		return
	}
	visited.Add(node.Pattern())
	checkNode(t, x, node)
	kids, err := node.Children()
	if err != nil {
		t.Fatal(err)
	}
	for _, kid := range kids {
		if !visited.Has(kid.Pattern()) {
			visit(t, x, visited, kid.(Node))
		}
	}
}

func checkNode(t testing.TB, x *assert.Assertions, node Node) {
	acount, err := node.AdjacentCount()
	if err != nil {
		t.Fatal(err)
	}
	kcount, err := node.ChildCount()
	if err != nil {
		t.Fatal(err)
	}
	kids, err := node.Children()
	if err != nil {
		t.Fatal(err)
	}
	pcount, err := node.ParentCount()
	if err != nil {
		t.Fatal(err)
	}
	parents, err := node.Parents()
	if err != nil {
		t.Fatal(err)
	}
	if kcount != len(kids) {
		x.Fail("kcount != len(kids)")
	}
	if pcount != len(parents) {
		x.Fail("count != len(parents)")
	}
	if kcount+pcount != acount {
		x.Fail("kcount + pcount != acount")
	}
	for _, kid := range kids {
		checkKid(t, x, node, kid.(Node))
	}
	for _, parent := range parents {
		checkKid(t, x, parent.(Node), node)
	}
}

func checkKid(t testing.TB, x *assert.Assertions, parent, kid Node) {
	pkids, err := parent.Children()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, pkid := range pkids {
		if bytes.Equal(pkid.Pattern().Label(), kid.Label()) {
			found = true
		}
	}
	if !found {
		t.Errorf("parent %v kids %v did not have %v", parent, pkids, kid)
		findChildren(parent, nil, true)
		t.Fatalf("assert-fail")
	}
	kparents, err := kid.Parents()
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, kparent := range kparents {
		if bytes.Equal(kparent.Pattern().Label(), parent.Label()) {
			found = true
		}
	}
	if !found {
		t.Fatalf("kid %v parents %v did not have %v", kid, kparents, parent)
	}
	pkids, err = parent.CanonKids()
	if err != nil {
		errors.Logf("ERROR", "err = %v", err)
		x.Fail("err != nil")
		os.Exit(2)
	}
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
		if err != nil {
			t.Fatal(err)
		}
		found = false
		for _, kparent := range kparents {
			if bytes.Equal(kparent.Pattern().Label(), parent.Label()) {
				continue
			}
			kparent_ckids, err := kparent.CanonKids()
			if err != nil {
				t.Fatal(err)
			}
			for _, kparent_ckid := range kparent_ckids {
				if bytes.Equal(kparent_ckid.Pattern().Label(), kid.Label()) {
					x.Fail(errors.Errorf("kid %v had multiple canon parents %v %v", kid, parent, kparent).Error())
				}
			}
		}
	}
}
