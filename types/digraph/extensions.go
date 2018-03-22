package digraph

import (
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/regrax/stats"
	"github.com/timtadh/regrax/types/digraph/digraph"
	"github.com/timtadh/regrax/types/digraph/subgraph"
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

func extensionPoint(G *digraph.Digraph, emb *subgraph.Embedding, e *digraph.Edge, src, targ int) *subgraph.Extension {
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

func validExtChecker(dt *Digraph, do func(*subgraph.Embedding, *subgraph.Extension)) func(*subgraph.Embedding, *digraph.Edge, int, int) int {
	return func(emb *subgraph.Embedding, e *digraph.Edge, src, targ int) int {
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

func extensionsFromEmbeddings(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool, seenMu *sync.Mutex) (total int, overlap []map[int]bool, fisEmbs []*subgraph.Embedding, sets []*hashtable.LinearHash, exts types.Set) {
	if dt.Mode&FIS == FIS || dt.Mode&GIS == GIS {
		if seen == nil {
			seen = make(map[int]bool)
		}
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
				seenMu.Lock()
				if seen[id] {
					seenIt = true
				}
				seenMu.Unlock()
			}
			if overlap != nil {
				if overlap[idx] == nil {
					overlap[idx] = make(map[int]bool)
				}
				overlap[idx][id] = true
			}
			seenMu.Lock()
			if seen != nil {
				seen[id] = true
			}
			seenMu.Unlock()
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
				add(emb, &dt.G.E[e], idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(emb, &dt.G.E[e], -1, idx)
			}
		}
		if fisEmbs != nil && !seenIt {
			fisEmbs = append(fisEmbs, emb)
		}
		total++
	}
	return total, overlap, fisEmbs, sets, exts
}

func extensionsFromFreqEdges(dt *Digraph, pattern *subgraph.SubGraph, ei subgraph.EmbIterator, seen map[int]bool, seenMu *sync.Mutex) (total int, overlap []map[int]bool, fisEmbs []*subgraph.Embedding, sets []*hashtable.LinearHash, exts types.Set) {
	if dt.Mode&FIS == FIS || dt.Mode&GIS == GIS {
		if seen == nil {
			seen = make(map[int]bool)
		}
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
			done <- hash
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
						exts <- ep
					}
				}
				ep := subgraph.NewExt(
					subgraph.Vertex{Idx: i, Color: u.Color},
					subgraph.Vertex{Idx: len(pattern.V), Color: e.TargColor},
					e.EdgeColor)
				exts <- ep
			}
			for _, e := range dt.Indices.EdgesToColor[u.Color] {
				ep := subgraph.NewExt(
					subgraph.Vertex{Idx: len(pattern.V), Color: e.SrcColor},
					subgraph.Vertex{Idx: i, Color: u.Color},
					e.EdgeColor)
				exts <- ep
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
				seenMu.Lock()
				if seen[id] {
					seenIt = true
				}
				seenMu.Unlock()
			}
			if overlap != nil {
				if overlap[idx] == nil {
					overlap[idx] = make(map[int]bool)
				}
				overlap[idx][id] = true
			}
			seenMu.Lock()
			if seen != nil {
				seen[id] = true
			}
			seenMu.Unlock()
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
func ExtsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph, patternOverlap []map[int]bool, unsupExts types.Set, mode Mode, debug bool) (int, []*subgraph.Extension, []*subgraph.Embedding, []map[int]bool, error) {
	// compute the embeddings
	var seen map[int]bool = nil
	var seenMu sync.Mutex
	var ei subgraph.EmbIterator
	switch {
	case mode&(MNI|FIS) != 0:
		ei = pattern.IterEmbeddings(
			dt.config.Workers(), dt.EmbSearchStartPoint, dt.Indices, patternOverlap, nil)
	case mode&(GIS) == GIS:
		seen = make(map[int]bool)
		ei = pattern.IterEmbeddings(
			dt.config.Workers(),
			dt.EmbSearchStartPoint,
			dt.Indices,
			patternOverlap,
			func(ids *subgraph.IdNode) bool {
				seenMu.Lock()
				defer seenMu.Unlock()
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
		return 0, nil, nil, nil, errors.Errorf("Unknown support counting strategy %v", mode)
	}

	// find the actual embeddings and compute the extensions
	// the extensions are stored in exts
	// the embeddings are stored in sets
	var exts types.Set
	var fisEmbs []*subgraph.Embedding
	var sets []*hashtable.LinearHash
	var overlap []map[int]bool
	var total int
	if mode&ExtFromEmb == ExtFromEmb && len(pattern.E) > 0 {
		// add the supported embeddings to the vertex sets
		// add the extensions to the extensions set
		total, overlap, fisEmbs, sets, exts = extensionsFromEmbeddings(dt, pattern, ei, seen, &seenMu)
		if total == 0 {
			// return 0, nil, nil, nil, nil, errors.Errorf("could not find any embedding of %v", pattern)
			// because we are extending from frequent edges for vertices this
			// is ok.
			return 0, nil, nil, nil, nil
		}
	} else if mode&ExtFromFreqEdges == ExtFromFreqEdges || len(pattern.E) <= 0 {
		total, overlap, fisEmbs, sets, exts = extensionsFromFreqEdges(dt, pattern, ei, seen, &seenMu)
		if total < dt.Support() {
			return 0, nil, nil, nil, nil
		}
	} else {
		return 0, nil, nil, nil, errors.Errorf("Unknown extension strategy %v", mode)
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

	var embeddings []*subgraph.Embedding
	if mode&(MNI) != 0 {
		// compute the minimally supported vertex
		arg, size := stats.Min(stats.RandomPermutation(len(sets)), func(i int) float64 {
			if sets[i] == nil {
				return 0
			}
			return float64(sets[i].Size())
		})
		if sets[arg] != nil {
			// construct the embeddings output slice
			embeddings = make([]*subgraph.Embedding, 0, int(size)+1)
			for i, next := sets[arg].Values()(); next != nil; i, next = next() {
				emb := i.(*subgraph.Embedding)
				embeddings = append(embeddings, emb)
			}
		}
	} else if mode&(FIS|GIS) != 0 {
		embeddings = fisEmbs
	} else {
		return 0, nil, nil, nil, errors.Errorf("Unknown support counting strategy %v", mode)
	}

	return len(embeddings), extensions, embeddings, overlap, nil
}
