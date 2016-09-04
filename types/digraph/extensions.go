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

func extensionsFromEmbeddings(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool) (total int, overlap []map[int]bool, fisEmbs []*subgraph.Embedding, sets []*hashtable.LinearHash, exts types.Set) {
	if dt.Mode&FIS == FIS {
		seen = make(map[int]bool)
		fisEmbs = make([]*subgraph.Embedding, 0, 10)
	} else {
		sets = make([]*hashtable.LinearHash, len(pattern.V))
	}
	if dt.Mode&OverlapPruning == OverlapPruning {
		overlap = make([]map[int]bool, len(pattern.V))
	}
	exts = set.NewSetMap(hashtable.NewLinearHash())
	add := validExtChecker(dt, func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		exts.Add(ext)
	})
	for emb, next := ei(false); next != nil; emb, next = next(false) {
		seenIt := false
		for idx, id := range emb.Ids {
			if fisEmbs != nil {
				if seen[id] {
					seenIt = true
				}
			}
			if overlap != nil {
				if overlap[idx] == nil {
					overlap[idx] = make(map[int]bool)
				}
				overlap[idx][id] = true
			}
			if seen != nil {
				seen[id] = true
			}
			if sets != nil {
				if sets[idx] == nil {
					sets[idx] = hashtable.NewLinearHash()
				}
				set := sets[idx]
				if !set.Has(types.Int(id)) {
					set.Put(types.Int(id), emb)
				}
			}
			for _, e := range dt.G.Kids[id] {
				add(emb, e, idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(emb, e, -1, idx)
			}
		}
		if fisEmbs != nil && !seenIt {
			fisEmbs = append(fisEmbs, emb)
		}
		total++
	}
	return total, overlap, fisEmbs, sets, exts
}

func extensionsFromFreqEdges(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool) (total int, overlap []map[int]bool, fisEmbs []*subgraph.Embedding, sets []*hashtable.LinearHash, exts types.Set) {
	if dt.Mode&FIS == FIS {
		seen = make(map[int]bool)
		fisEmbs = make([]*subgraph.Embedding, 0, 10)
	} else {
		sets = make([]*hashtable.LinearHash, len(pattern.V))
	}
	if dt.Mode&OverlapPruning == OverlapPruning {
		overlap = make([]map[int]bool, len(pattern.V))
	}
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
	stop := false
	for emb, next := ei(stop); next != nil; emb, next = next(stop) {
		min := -1
		seenIt := false
		for idx, id := range emb.Ids {
			if fisEmbs != nil {
				if seen[id] {
					seenIt = true
				}
			}
			if overlap != nil {
				if overlap[idx] == nil {
					overlap[idx] = make(map[int]bool)
				}
				overlap[idx][id] = true
			}
			if seen != nil {
				seen[id] = true
			}
			if sets != nil {
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
		}
		if fisEmbs != nil && !seenIt {
			fisEmbs = append(fisEmbs, emb)
			min = len(fisEmbs)
		}
		total++
		if min >= support {
			stop = true
		}
	}
	if total < support {
		return total, overlap, fisEmbs, sets, nil
	}
	return total, overlap, fisEmbs, sets, <-done
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
	var ei subgraph.EmbIterator
	var dropped *subgraph.VertexEmbeddings
	switch {
	case mode&(MNI|FIS) != 0:
		ei, dropped = pattern.IterEmbeddings(
			dt.EmbSearchStartPoint, dt.Indices, unsupEmbs, patternOverlap, nil)
	case mode&(GIS) == GIS:
		seen = make(map[int]bool)
		ei, dropped = pattern.IterEmbeddings(
			dt.EmbSearchStartPoint,
			dt.Indices,
			unsupEmbs,
			patternOverlap,
			func(ids *subgraph.IdNode) bool {
				for c := ids; c != nil; c = c.Prev {
					if _, has := seen[c.Id]; has {
						for c := ids; c != nil; c = c.Prev {
							seen[c.Id] = true
						}
						return true
					}
				}
				return false
			})
	default:
		return 0, nil, nil, nil, nil, errors.Errorf("Unknown support counting strategy %v", mode)
	}

	// find the actual embeddings and compute the extensions
	// the extensions are stored in exts
	// the embeddings are stored in sets
	var exts types.Set
	var fisEmbs []*subgraph.Embedding
	var sets []*hashtable.LinearHash
	var overlap []map[int]bool
	var total int
	if mode&ExtFromEmb == ExtFromEmb {
		// add the supported embeddings to the vertex sets
		// add the extensions to the extensions set
		total, overlap, fisEmbs, sets, exts = extensionsFromEmbeddings(dt, pattern, ei, seen)
		if total == 0 {
			return 0, nil, nil, nil, nil, errors.Errorf("could not find any embedding of %v", pattern)
		}
	} else if mode&ExtFromFreqEdges == ExtFromFreqEdges {
		total, overlap, fisEmbs, sets, exts = extensionsFromFreqEdges(dt, pattern, ei, seen)
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

	if mode&EmbeddingPruning == EmbeddingPruning && unsupEmbs != nil {
		// for i, next := unsupEmbs.Items()(); next != nil; i, next = next() {
		for emb, ok := range unsupEmbs {
			if ok {
				*dropped = append(*dropped, &emb)
			}
		}
	}

	var embeddings []*subgraph.Embedding
	if mode&(MNI|GIS) != 0 {
		// compute the minimally supported vertex
		arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
			return float64(sets[i].Size())
		})
		// construct the embeddings output slice
		embeddings = make([]*subgraph.Embedding, 0, int(size)+1)
		for i, next := sets[arg].Values()(); next != nil; i, next = next() {
			emb := i.(*subgraph.Embedding)
			embeddings = append(embeddings, emb)
		}
	} else if mode&(FIS) == FIS {
		embeddings = fisEmbs
	} else {
		return 0, nil, nil, nil, nil, errors.Errorf("Unknown support counting strategy %v", mode)
	}

	if CACHE_DEBUG || debug {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v total-embs %v : %v", len(extensions), len(embeddings), total, pattern.Pretty(dt.G.Colors))
	}
	if !debug {
		err := cacheExtsEmbs(dt, pattern, len(embeddings), extensions, embeddings, overlap, *dropped)
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
