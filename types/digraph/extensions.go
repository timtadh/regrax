package digraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
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
//    same pattern. Memoization needs to be added! (done)
//
// 2. There may be duplicate embeddings computed. Investigate. (recheck)
//
// 3. There may be automorphic embeddings computed. Investigate. (recheck)
//
// 4. Instead of the full Embeddings we could work in overlap space. (todo)
//    Investigate.

func extensionPoint(G *goiso.Graph, emb *subgraph.Embedding, e *goiso.Edge) *subgraph.Extension {
	hasTarg := false
	hasSrc := false
	var srcIdx int = len(emb.SG.V)
	var targIdx int = len(emb.SG.V)
	for idx, id := range emb.Ids {
		if e.Src == id {
			hasSrc = true
			srcIdx = idx
		}
		if e.Targ == id {
			hasTarg = true
			targIdx = idx
		}
	}
	if !hasTarg && !hasSrc {
		srcIdx = len(emb.SG.V)
		targIdx = len(emb.SG.V) + 1
	}
	src := subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color}
	targ := subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color}
	ext := subgraph.NewExt(src, targ, e.Color)
	return ext
}

func validExtChecker(dt *Digraph, do func(*subgraph.Embedding, *subgraph.Extension)) func(*subgraph.Embedding, *goiso.Edge) int {
	return func(emb *subgraph.Embedding, e *goiso.Edge) int {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return 0
		}
		ep := extensionPoint(dt.G, emb, e)
		if !emb.HasExtension(ep) {
			do(emb, ep)
			return 1
		}
		return 0
	}
}

// unique extensions
func extensions(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, error) {
	// compute the embeddings
	ei, err := pattern.IterEmbeddings(dt.G, dt.Indices, nil)
	if err != nil {
		return nil, err
	}
	exts := set.NewSortedSet(10)
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ep *subgraph.Extension) {
		exts.Add(ep)
	})

	// add the extensions to the extensions set
	for emb, next := ei(); next != nil; emb, next = next() {
		for i := range emb.SG.V {
			id := emb.Ids[i]
			for _, e := range dt.G.Kids[id] {
				add(emb, e)
			}
			for _, e := range dt.G.Parents[id] {
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
func extsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, []*subgraph.Embedding, error) {
	if has, exts, embs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
		return nil, nil, err
	} else if has {
		if false {
			errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v", pattern, len(exts), len(embs))
		}
		return exts, embs, nil
	}
	// errors.Logf("DEBUG", "----   extsAndEmbs pattern %v", pattern)
	// compute the embeddings
	ei, err := pattern.IterEmbeddings(dt.G, dt.Indices, nil)
	if err != nil {
		return nil, nil, err
	}
	exts := hashtable.NewLinearHash()
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		exts.Put(ext, nil)
	})
	sets := make([]*set.MapSet, len(pattern.V))

	total := 0
	// add the supported embeddings to the vertex sets
	// add the extensions to the extensions set
	// errors.Logf("DEBUG", "parent %v", parent)
	// errors.Logf("DEBUG", "computing embeddings %v", pattern.Pretty(dt.G.Colors))
	// errors.Logf("DEBUG", "computing embeddings %v", pattern)
	for emb, next := ei(); next != nil; emb, next = next() {
		// errors.Logf("DEBUG", "emb %v", emb)
		for idx, id := range emb.Ids {
			if sets[idx] == nil {
				sets[idx] = set.NewMapSet(set.NewSortedSet(10))
			}
			set := sets[idx]
			if !set.Has(types.Int(id)) {
				set.Put(types.Int(id), emb)
			}
			for _, e := range dt.G.Kids[id] {
				add(emb, e)
			}
			for _, e := range dt.G.Parents[id] {
				add(emb, e)
			}
		}
		total++
	}

	if total == 0 {
		return nil, nil, errors.Errorf("could not find any embedding of %v", pattern)
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Keys()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		extensions = append(extensions, ext)
	}
	// errors.Logf("DEBUG", "----   exts %v", exts)

	// errors.Logf("DEBUG", "pattern %v len sets %v sets %v", pattern, len(sets), sets)
	// compute the minimally supported vertex
	arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})
	// errors.Logf("DEBUG", "----   support %v", size)

	// construct the embeddings output slice
	embeddings := make([]*subgraph.Embedding, 0, int(size)+1)
	for i, next := sets[arg].Values()(); next != nil; i, next = next() {
		emb := i.(*subgraph.Embedding)
		embeddings = append(embeddings, emb)
	}

	// errors.Logf("DEBUG", "pat %v total-embeddings %v supported %v unique-ext %v", pattern, total, len(embeddings), len(extensions))

	// return it all
	if true {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v total-embs %v : %v", len(extensions), len(embeddings), total, pattern.Pretty(dt.G.Colors))
	}
	err = cacheExtsEmbs(dt, pattern, extensions, embeddings)
	if err != nil {
		return nil, nil, err
	}
	return extensions, embeddings, nil
}

func cacheExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph, exts []*subgraph.Extension, embs []*subgraph.Embedding) error {
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
	for _, emb := range embs {
		err := dt.Embeddings.Add(emb.SG, emb)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadCachedExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph) (bool, []*subgraph.Extension, []*subgraph.Embedding, error) {
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

	embs := make([]*subgraph.Embedding, 0, 10)
	err = dt.Embeddings.DoFind(pattern, func(_ *subgraph.SubGraph, emb *subgraph.Embedding) error {
		embs = append(embs, emb)
		return nil
	})
	if err != nil {
		return false, nil, nil, err
	}

	return true, exts, embs, nil
}
