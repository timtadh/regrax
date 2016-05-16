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

const CACHE_DEBUG = false

type Mode uint8

const (
	Automorphs = iota
	NoAutomorphs
	OptimisticPruning
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

func extensionPoint(G *goiso.Graph, emb *subgraph.Embedding, e *goiso.Edge, src, targ int) *subgraph.Extension {
	hasTarg := false
	hasSrc := false
	var srcIdx int = len(emb.SG.V)
	var targIdx int = len(emb.SG.V)
	if src >= 0 {
		hasSrc = true
		srcIdx = src
	}
	if targ >= 0 {
		hasTarg = true
		targIdx = targ
	}
	for idx, id := range emb.Ids {
		if !hasSrc && e.Src == id {
			hasSrc = true
			srcIdx = idx
		}
		if !hasTarg && e.Targ == id {
			hasTarg = true
			targIdx = idx
		}
		if hasTarg && hasSrc {
			break
		}
	}
	return subgraph.NewExt(
		subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color},
		subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color},
		e.Color)
}

func validExtChecker(dt *Digraph, do func(*subgraph.Embedding, *subgraph.Extension)) func(*subgraph.Embedding, *goiso.Edge, int, int) int {
	return func(emb *subgraph.Embedding, e *goiso.Edge, src, targ int) int {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return 0
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return 0
		}
		ep := extensionPoint(dt.G, emb, e, src, targ)
		if !emb.SG.HasExtension(ep) {
			do(emb, ep)
			return 1
		}
		return 0
	}
}

// unique extensions
func extensions(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, error) {
	// compute the embeddings
	ei, err := pattern.IterEmbeddings(dt.Indices, nil)
	if err != nil {
		return nil, err
	}
	exts := set.NewSortedSet(10)
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ep *subgraph.Extension) {
		exts.Add(ep)
	})

	// add the extensions to the extensions set
	for emb, next := ei(); next != nil; emb, next = next() {
		for idx := range emb.SG.V {
			id := emb.Ids[idx]
			for _, e := range dt.G.Kids[id] {
				add(emb, e, idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(emb, e, -1, idx)
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
func ExtsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph, unsupported types.Set, mode Mode, debug bool) (int, []*subgraph.Extension, []*subgraph.Embedding, error) {
	if !debug {
		if has, support, exts, embs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
			return 0, nil, nil, err
		} else if has {
			if false {
				errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v", pattern, len(exts), len(embs))
			}
			return support, exts, embs, nil
		}
	}
	if CACHE_DEBUG || debug {
		errors.Logf("CACHE-DEBUG", "ExtsAndEmbs %v", pattern.Pretty(dt.G.Colors))
	}
	// compute the embeddings
	var seen map[int]bool = nil
	var err error
	var ei subgraph.EmbIterator
	switch mode {
	case Automorphs:
		ei, err = pattern.IterEmbeddings(dt.Indices, nil)
	case NoAutomorphs:
		ei, err = subgraph.FilterAutomorphs(pattern.IterEmbeddings(dt.Indices, nil))
	case OptimisticPruning:
		seen = make(map[int]bool)
		ei, err = pattern.IterEmbeddings(
			dt.Indices,
			func(ids *subgraph.IdNode) bool {
				for c := ids; c != nil; c = c.Prev {
					if _, has := seen[c.Id]; !has {
						return false
					}
				}
				return true
		})
	default:
		return 0, nil, nil, errors.Errorf("Unknown mode %v", mode)
	}
	if err != nil {
			return 0, nil, nil, err
	}
	exts := set.NewSetMap(hashtable.NewLinearHash())
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		if unsupported.Has(ext) {
			return
		}
		exts.Add(ext)
	})
	sets := make([]*hashtable.LinearHash, len(pattern.V))

	total := 0
	// add the supported embeddings to the vertex sets
	// add the extensions to the extensions set
	// errors.Logf("DEBUG", "parent %v", parent)
	// errors.Logf("DEBUG", "computing embeddings %v", pattern.Pretty(dt.G.Colors))
	// errors.Logf("DEBUG", "computing embeddings %v", pattern)
	for emb, next := ei(); next != nil; emb, next = next() {
		// errors.Logf("DEBUG", "emb %v", emb)
		for idx, id := range emb.Ids {
			if seen != nil {
				seen[id] = true
			}
			if sets[idx] == nil {
				sets[idx] = hashtable.NewLinearHash()
			}
			set := sets[idx]
			if !set.Has(types.Int(id)) {
				set.Put(types.Int(id), emb)
			}
			for _, e := range dt.G.Kids[id] {
				add(emb, e, idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(emb, e, -1, idx)
			}
		}
		total++
		// const limit = 10000
		// if total > limit {
		// 	errors.Logf("WARNING", "skipping the rest of the embeddings for %v (over %v)", pattern, limit)
		// 	break
		// }
	}

	if total == 0 {
		return 0, nil, nil, errors.Errorf("could not find any embedding of %v", pattern)
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Items()(); next != nil; i, next = next() {
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
	if CACHE_DEBUG || debug {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v total-embs %v : %v", len(extensions), len(embeddings), total, pattern.Pretty(dt.G.Colors))
	}
	if !debug {
		err = cacheExtsEmbs(dt, pattern, len(embeddings), extensions, embeddings)
		if err != nil {
			return 0, nil, nil, err
		}
	}
	return len(embeddings), extensions, embeddings, nil
}

func cacheExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph, support int, exts []*subgraph.Extension, embs []*subgraph.Embedding) error {
	label := pattern.Label()
	if has, err := dt.Frequency.Has(label); err != nil {
		return err
	} else if has {
		return nil
	}
	err := dt.Frequency.Add(label, int32(support))
	if err != nil {
		return nil
	}
	if support < dt.Support() {
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

func loadCachedExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph) (bool, int, []*subgraph.Extension, []*subgraph.Embedding, error) {
	label := pattern.Label()
	if has, err := dt.Frequency.Has(label); err != nil {
		return false, 0, nil, nil, err
	} else if !has {
		return false, 0, nil, nil, nil
	}

	support := 0
	err := dt.Frequency.DoFind(label, func(_ []byte, s int32) error {
		support = int(s)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, err
	}

	exts := make([]*subgraph.Extension, 0, 10)
	err = dt.Extensions.DoFind(label, func(_ []byte, ext *subgraph.Extension) error {
		exts = append(exts, ext)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, err
	}

	embs := make([]*subgraph.Embedding, 0, 10)
	err = dt.Embeddings.DoFind(pattern, func(_ *subgraph.SubGraph, emb *subgraph.Embedding) error {
		embs = append(embs, emb)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, err
	}

	return true, support, exts, embs, nil
}
