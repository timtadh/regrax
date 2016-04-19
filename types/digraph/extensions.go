package digraph

import (
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
	// "github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stats"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

// YOU ARE HERE:
//
// Ok so what we need to do is take an a pattern (eg. subgraph.SubGraph) and
// compute all of the unique subgraph.Extension(s). Those will then get stored
// in digraph.Extension.
//
// Then when computing children the extensions are looked up rather than found
// using the embeddings list. Furthermore, the only supported embeddings are
// stored in the embedding list to cut down on space.
//
// kk.


func extensionPoint(G *goiso.Graph, sg *goiso.SubGraph, e *goiso.Edge) *subgraph.Extension {
	hasTarg := false
	hasSrc := false
	var srcIdx int = len(sg.V)
	var targIdx int = len(sg.V)
	for _, v := range sg.V {
		if e.Src == v.Id {
			hasSrc = true
			srcIdx = v.Idx
		}
		if e.Targ == v.Id {
			hasTarg = true
			targIdx = v.Idx
		}
	}
	if !hasTarg && !hasSrc {
		srcIdx = 0
		targIdx = 1
	}
	src := subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color}
	targ := subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color}
	ext := subgraph.NewExt(src, targ, e.Color)
	return ext
}


// unique extensions
func extensions(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, error) {
	// compute the embeddings
	ei, err := pattern.IterEmbeddings(dt.G, dt.ColorMap, dt.Extender, nil)
	if err != nil {
		return nil, err
	}
	exts := set.NewSortedSet(10)
	add := validExtChecker(dt, func(sg *goiso.SubGraph, e *goiso.Edge) {
		exts.Add(extensionPoint(dt.G, sg, e))
	})

	// add the extensions to the extensions set
	for emb, next := ei(); next != nil; emb, next = next() {
		for i := range emb.V {
			u := &emb.V[i]
			for _, e := range dt.G.Kids[u.Id] {
				add(emb, e)
			}
			for _, e := range dt.G.Parents[u.Id] {
				add(emb, e)
			}
		}
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Items()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		extensions = append(extensions, ext)
	}

	return extensions, nil
}

// unique extensions and supported embeddings
func extsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, []*goiso.SubGraph, error) {
	// errors.Logf("DEBUG", "----   extsAndEmbs pattern %v", pattern)
	// compute the embeddings
	ei, err := pattern.IterEmbeddings(dt.G, dt.ColorMap, dt.Extender, nil)
	if err != nil {
		return nil, nil, err
	}
	exts := hashtable.NewLinearHash()
	add := validExtChecker(dt, func(sg *goiso.SubGraph, e *goiso.Edge) {
		ep := extensionPoint(dt.G, sg, e)
		exts.Put(ep, nil)
		// errors.Logf("DEBUG", "----   -- ext point %v", ep)
	})
	sets := make([]*set.MapSet, len(pattern.V))

	total := 0
	// add the supported embeddings to the vertex sets
	// add the extensions to the extensions set
	for emb, next := ei(); next != nil; emb, next = next() {
		// errors.Logf("DEBUG", "----   - emb %v", emb.Label())
		for i := range emb.V {
			if sets[i] == nil {
				sets[i] = set.NewMapSet(set.NewSortedSet(10))
			}
			set := sets[i]
			u := &emb.V[i]
			id := types.Int(u.Id)
			if !set.Has(id) {
				set.Put(id, emb)
			}
			for _, e := range dt.G.Kids[u.Id] {
				add(emb, e)
			}
			for _, e := range dt.G.Parents[u.Id] {
				add(emb, e)
			}
		}
		total++
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Keys()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		extensions = append(extensions, ext)
	}
	// errors.Logf("DEBUG", "----   exts %v", exts)
	
	// compute the minimally supported vertex
	arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	// errors.Logf("DEBUG", "----   support %v", size)

	// construct the embeddings output slice
	embeddings := make([]*goiso.SubGraph, 0, int(size)+1)
	for i, next := sets[arg].Values()(); next != nil; i, next = next() {
		emb := i.(*goiso.SubGraph)
		embeddings = append(embeddings, emb)
	}

	// errors.Logf("DEBUG", "pat %v total-embeddings %v supported %v unique-ext %v", pattern, total, len(embeddings), len(extensions))

	// return it all
	return extensions, embeddings, nil
}

