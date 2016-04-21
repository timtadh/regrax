package digraph

import (
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stats"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

// YOU ARE HERE:
//
// 1. The embeddings and extensions are being computed multiple times for the
//    same pattern. Memoization needs to be added!
//
// 2. There may be duplicate embeddings computed. Investigate.
//
// 3. There may be automorphic embeddings computed. Investigate.
//
// 4. Instead of the full Embeddings we could work in overlap space.
//    Investigate.


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
	if has, exts, embs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
		return nil, nil, err
	} else if has {
		errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v", pattern, len(exts), len(embs))
		return exts, embs, nil
	}
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
	return extensions, embeddings, cacheExtsEmbs(dt, pattern, extensions, embeddings)
}

func cacheExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*goiso.SubGraph) error {
	label := pattern.Label()
	if has, err := dt.Frequency.Has(label); err != nil {
		return err
	} else if has {
		return nil
	}
	err := dt.Frequency.Add(label, int32(len(embs)))
	if err != nil {
		return nil
	}
	if len(embs) < dt.Support() {
		return nil
	}
	// save the supported extensions and embeddings
	for _, ext := range exts {
		err := dt.Extensions.Add(label, ext)
		if err != nil {
			return err
		}
	}
	for _, sg := range embs {
		err := dt.Embeddings.Add(label, sg)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadCachedExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph) (bool, []*subgraph.Extension, []*goiso.SubGraph, error) {
	label := pattern.Label()
	if has, err := dt.Frequency.Has(label); err != nil {
		return false, nil, nil, err
	} else if !has {
		return false, nil, nil, nil
	}

	exts := make([]*subgraph.Extension, 0, 10)
	err := dt.Extensions.DoFind(label, func(_ []byte, ext *subgraph.Extension) error {
		exts = append(exts, ext)
		return nil
	})
	if err != nil {
		return false, nil, nil, err
	}

	embs := make([]*goiso.SubGraph, 0, 10)
	err = dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
		embs = append(embs, sg)
		return nil
	})
	if err != nil {
		return false, nil, nil, err
	}

	return true, exts, embs, nil
}


