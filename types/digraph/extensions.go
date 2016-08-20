package digraph

import ()

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

// YOU ARE HERE:
//
// 1. The embeddings and extensions are being computed multiple times for the
//    same pattern. Memoization needs to be added! (done)
//
// 2. There may be duplicate embeddings computed. Investigate. (does not happen)
//
// 3. There may be automorphic embeddings computed. Investigate. (happens)
//
// 4. Instead of the full Embeddings we could work in overlap space.
//    Investigate. (done, it was difficult and not worth it)
//
// 5. Add a parallel implementation of extending from embedding list ala original
//    graple. This will give a benchmarking point of comparison.

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
		if hasTarg && hasSrc {
			break
		}
		if !hasSrc && e.Src == id {
			hasSrc = true
			srcIdx = idx
		}
		if !hasTarg && e.Targ == id {
			hasTarg = true
			targIdx = idx
		}
	}
	return subgraph.NewExt(
		subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color},
		subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color},
		e.Color)
}

func validExtChecker(dt *Digraph, do func(*subgraph.Embedding, *subgraph.Extension)) func(*subgraph.Embedding, *goiso.Edge, int, int) int {
	return func(emb *subgraph.Embedding, e *goiso.Edge, src, targ int) int {
		if dt.Indices.EdgeCounts[dt.Indices.Colors(e)] < dt.Support() {
			return 0
		}
		// if dt.G.ColorFrequency(e.Color) < dt.Support() {
		// 	return 0
		// } else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
		// 	return 0
		// } else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
		// 	return 0
		// }
		ep := extensionPoint(dt.G, emb, e, src, targ)
		if !emb.SG.HasExtension(ep) {
			do(emb, ep)
			return 1
		}
		return 0
	}
}

func extensionsFromEmbeddings(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool) (total int, sets []*hashtable.LinearHash, exts types.Set) {
	exts = set.NewSetMap(hashtable.NewLinearHash())
	sets = make([]*hashtable.LinearHash, len(pattern.V))
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		exts.Add(ext)
	})
	for emb, next := ei(); next != nil; emb, next = next() {
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
	}
	return total, sets, exts
}

func extensionsFromFreqEdges(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool) (total int, sets []*hashtable.LinearHash, exts types.Set) {
	sets = make([]*hashtable.LinearHash, len(pattern.V))
	support := dt.Support()
	done := make(chan types.Set)
	go func(done chan types.Set) {
		exts := make(chan *subgraph.Extension, len(pattern.V))
		go func() {
			hash := set.NewSetMap(hashtable.NewLinearHash())
			for ext := range exts {
				if !pattern.HasExtension(ext) {
					hash.Add(ext)
				}
			}
			done<-hash
			close(done)
		}()
		for i := range pattern.V {
			u := &pattern.V[i]
			for _, e := range dt.Indices.EdgesFromColor[u.Color] {
				for j := range pattern.V {
					v := &pattern.V[j]
					if v.Color == e.TargColor {
						ep := subgraph.NewExt(
							subgraph.Vertex{Idx: i, Color: e.SrcColor},
							subgraph.Vertex{Idx: j, Color: e.TargColor},
							e.EdgeColor)
						exts <-ep
					}
				}
				ep := subgraph.NewExt(
					subgraph.Vertex{Idx: i, Color: u.Color},
					subgraph.Vertex{Idx: len(pattern.V), Color: e.TargColor},
					e.EdgeColor)
				exts <-ep
			}
			for _, e := range dt.Indices.EdgesToColor[u.Color] {
				ep := subgraph.NewExt(
					subgraph.Vertex{Idx: len(pattern.V), Color: e.SrcColor},
					subgraph.Vertex{Idx: i, Color: u.Color},
					e.EdgeColor)
				exts <-ep
			}
		}
		close(exts)
	}(done)
	for emb, next := ei(); next != nil; emb, next = next() {
		min := -1
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
			size := set.Size()
			if min == -1 || size < min {
				min = size
			}
		}
		total++
		if min >= support {
			break
		}
	}
	if total < support {
		return total, sets, nil
	}
	return total, sets, <-done
}

// unique extensions and supported embeddings
func ExtsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph, patternOverlap []map[int]bool, unsupExts types.Set, unsupEmbs map[subgraph.VrtEmb]bool, mode Mode, debug bool) (int, []*subgraph.Extension, []*subgraph.Embedding, []map[int]bool, subgraph.VertexEmbeddings, error) {
	if !debug {
		if has, support, exts, embs, overlap, unsupEmbs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
			return 0, nil, nil, nil, nil, err
		} else if has {
			if false {
				errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v", pattern, len(exts), len(embs))
			}
			return support, exts, embs, overlap, unsupEmbs, nil
		}
	}
	if CACHE_DEBUG || debug {
		errors.Logf("CACHE-DEBUG", "ExtsAndEmbs %v", pattern.Pretty(dt.G.Colors))
	}

	// compute the embeddings
	var seen map[int]bool = nil
	var err error
	var ei subgraph.EmbIterator
	var dropped *subgraph.VertexEmbeddings
	switch {
	case mode&Automorphs == Automorphs:
		ei, dropped, err = pattern.IterEmbeddings(dt.Indices, unsupEmbs, patternOverlap, nil)
	case mode&NoAutomorphs == NoAutomorphs:
		ei, dropped, err = subgraph.FilterAutomorphs(pattern.IterEmbeddings(dt.Indices, unsupEmbs, patternOverlap, nil))
	case mode&OptimisticPruning == OptimisticPruning:
		seen = make(map[int]bool)
		ei, dropped, err = pattern.IterEmbeddings(
			dt.Indices,
			unsupEmbs,
			patternOverlap,
			func(ids *subgraph.IdNode) bool {
				for c := ids; c != nil; c = c.Prev {
					if _, has := seen[c.Id]; !has {
						return false
					}
				}
				return true
			})
	default:
		return 0, nil, nil, nil, nil, errors.Errorf("Unknown embedding search strategy %v", mode)
	}
	if err != nil {
		return 0, nil, nil, nil, nil, err
	}

	// find the actual embeddings and compute the extensions
	// the extensions are stored in exts
	// the embeddings are stored in sets
	var exts types.Set
	var sets []*hashtable.LinearHash
	var total int
	if mode&ExtFromEmb == ExtFromEmb {
		// add the supported embeddings to the vertex sets
		// add the extensions to the extensions set
		total, sets, exts = extensionsFromEmbeddings(dt, pattern, ei, seen)
		if total == 0 {
			return 0, nil, nil, nil, nil, errors.Errorf("could not find any embedding of %v", pattern)
		}
	} else if mode&ExtFromFreqEdges == ExtFromFreqEdges {
		total, sets, exts = extensionsFromFreqEdges(dt, pattern, ei, seen)
		if total < dt.Support() {
			return 0, nil, nil, nil, nil, nil
		}
	} else {
		return 0, nil, nil, nil, nil, errors.Errorf("Unknown extension strategy %v", mode)
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Items()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		if unsupExts != nil && unsupExts.Has(ext) {
			continue
		}
		extensions = append(extensions, ext)
	}

	// construct the overlap
	var overlap []map[int]bool = nil
	if dt.Overlap != nil {
		overlap = make([]map[int]bool, 0, len(pattern.V))
		for _, s := range sets {
			o := make(map[int]bool, s.Size())
			for x, next := s.Keys()(); next != nil; x, next = next() {
				o[int(x.(types.Int))] = true
			}
			overlap = append(overlap, o)
		}
	}

	if mode&EmbeddingPruning == EmbeddingPruning && unsupEmbs != nil {
		//for i, next := unsupEmbs.Items()(); next != nil; i, next = next() {
		// for emb, ok := range unsupEmbs {
		// 	if ok {
		// 		*dropped = append(*dropped, &emb)
		// 	}
		// }
	}

	// compute the minimally supported vertex
	arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
		return float64(sets[i].Size())
	})

	// construct the embeddings output slice
	embeddings := make([]*subgraph.Embedding, 0, int(size)+1)
	for i, next := sets[arg].Values()(); next != nil; i, next = next() {
		emb := i.(*subgraph.Embedding)
		embeddings = append(embeddings, emb)
	}

	if CACHE_DEBUG || debug {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v total-embs %v : %v", len(extensions), len(embeddings), total, pattern.Pretty(dt.G.Colors))
	}
	if !debug {
		err = cacheExtsEmbs(dt, pattern, len(embeddings), extensions, embeddings, overlap, *dropped)
		if err != nil {
			return 0, nil, nil, nil, nil, err
		}
	}
	return len(embeddings), extensions, embeddings, overlap, *dropped, nil
}

func cacheExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph, support int, exts []*subgraph.Extension, embs []*subgraph.Embedding, overlap []map[int]bool, unsupEmbs subgraph.VertexEmbeddings) error {
	if dt.Mode & Caching == 0 && len(pattern.E) > 0 {
		return nil
	}
	dt.lock.Lock()
	defer dt.lock.Unlock()
	label := pattern.Label()
	// frequency will always get added, so if frequency has the label
	// this pattern has already been saved
	if has, err := dt.Frequency.Has(label); err != nil {
		return err
	} else if has {
		return nil
	}
	err := dt.Frequency.Add(label, int32(support))
	if err != nil {
		return nil
	}
	// if the support is too low we can bail on saving the rest of the
	// node
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
	if dt.Overlap != nil && len(pattern.E) > 3 {
		// save the overlap if using
		err = dt.Overlap.Add(pattern, overlap)
		if err != nil {
			return err
		}
	}
	if dt.UnsupEmbs != nil {
		// for _, emb := range unsupEmbs {
		// 	err = dt.UnsupEmbs.Add(pattern, emb)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}
	return nil
}

func loadCachedExtsEmbs(dt *Digraph, pattern *subgraph.SubGraph) (bool, int, []*subgraph.Extension, []*subgraph.Embedding, []map[int]bool, subgraph.VertexEmbeddings, error) {
	if dt.Mode & Caching == 0 && len(pattern.E) > 0 {
		return false, 0, nil, nil, nil, nil, nil
	}
	dt.lock.RLock()
	defer dt.lock.RUnlock()
	label := pattern.Label()
	if has, err := dt.Frequency.Has(label); err != nil {
		return false, 0, nil, nil, nil, nil, err
	} else if !has {
		return false, 0, nil, nil, nil, nil, nil
	}

	support := 0
	err := dt.Frequency.DoFind(label, func(_ []byte, s int32) error {
		support = int(s)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, nil, nil, err
	}

	exts := make([]*subgraph.Extension, 0, 10)
	err = dt.Extensions.DoFind(label, func(_ []byte, ext *subgraph.Extension) error {
		exts = append(exts, ext)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, nil, nil, err
	}

	embs := make([]*subgraph.Embedding, 0, 10)
	err = dt.Embeddings.DoFind(pattern, func(_ *subgraph.SubGraph, emb *subgraph.Embedding) error {
		embs = append(embs, emb)
		return nil
	})
	if err != nil {
		return false, 0, nil, nil, nil, nil, err
	}

	var overlap []map[int]bool = nil
	if dt.Overlap != nil {
		err = dt.Overlap.DoFind(pattern, func(_ *subgraph.SubGraph, o []map[int]bool) error {
			overlap = o
			return nil
		})
		if err != nil {
			return false, 0, nil, nil, nil, nil, err
		}
	}
	var unsupEmbs subgraph.VertexEmbeddings = nil
	if dt.UnsupEmbs != nil {
		// unsupEmbs = make([]*subgraph.Embedding, 0, 10)
		// err = dt.Embeddings.DoFind(pattern, func(_ *subgraph.SubGraph, emb *subgraph.Embedding) error {
		// 	unsupEmbs = append(unsupEmbs, emb)
		// 	return nil
		// })
		// if err != nil {
		// 	return false, 0, nil, nil, nil, nil, err
		// }
	}

	return true, support, exts, embs, overlap, unsupEmbs, nil
}
