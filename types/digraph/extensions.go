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

func overlapExtensionPoints(G *goiso.Graph, o *subgraph.Overlap, e *goiso.Edge, src, targ int) []*subgraph.Extension {
	exts := make([]*subgraph.Extension, 0, 10)
	if src >= 0 {
		srcIdx := src
		targs := make([]int, 0, len(o.SG.V))
		for idx, ids := range o.Ids {
			if ids.Has(types.Int(e.Targ)) {
				targs = append(targs, idx)
			}
		}
		if len(targs) == 0 {
			exts = append(exts, subgraph.NewExt(
				subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color},
				subgraph.Vertex{Idx: len(o.SG.V), Color: G.V[e.Targ].Color},
				e.Color,
			))
		}
		for _, targIdx := range targs {
			exts = append(exts, subgraph.NewExt(
				subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color},
				subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color},
				e.Color,
			))
		}
	} else if targ >= 0 {
		targIdx := targ
		srcs := make([]int, 0, len(o.SG.V))
		for idx, ids := range o.Ids {
			if ids.Has(types.Int(e.Src)) {
				srcs = append(srcs, idx)
			}
		}
		if len(srcs) == 0 {
			exts = append(exts, subgraph.NewExt(
				subgraph.Vertex{Idx: len(o.SG.V), Color: G.V[e.Src].Color},
				subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color},
				e.Color,
			))
		}
		for _, srcIdx := range srcs {
			exts = append(exts, subgraph.NewExt(
				subgraph.Vertex{Idx: srcIdx, Color: G.V[e.Src].Color},
				subgraph.Vertex{Idx: targIdx, Color: G.V[e.Targ].Color},
				e.Color,
			))
		}
	} else {
		panic("unreachable")
	}
	return exts
}

func overlapValidExtChecker(dt *Digraph, do func(*subgraph.Overlap, *subgraph.Extension)) func(*subgraph.Overlap, *goiso.Edge, int, int) {
	support := dt.Support()
	return func(o *subgraph.Overlap, e *goiso.Edge, src, targ int) {
		if dt.G.ColorFrequency(e.Color) < support {
			return
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < support {
			return
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < support {
			return
		}
		for _, ep := range overlapExtensionPoints(dt.G, o, e, src, targ) {
			if !o.SG.HasExtension(ep) {
				do(o, ep)
			}
		}
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

var ExtsAndEmbs func(dt *Digraph, pattern *subgraph.SubGraph, unsupported types.Set) (int, []*subgraph.Extension, []*subgraph.Embedding, error) = extsAndEmbs_1

// unique extensions and supported embeddings
func extsAndEmbs_1(dt *Digraph, pattern *subgraph.SubGraph, unsupported types.Set) (int, []*subgraph.Extension, []*subgraph.Embedding, error) {
	if has, support, exts, embs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
		return 0, nil, nil, err
	} else if has {
		if false {
			errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v", pattern, len(exts), len(embs))
		}
		return support, exts, embs, nil
	}
	// errors.Logf("DEBUG", "----   extsAndEmbs pattern %v", pattern)
	// compute the embeddings
	seen := make(map[int]bool)
	ei, err := subgraph.FilterAutomorphs(pattern.IterEmbeddings(
		dt.Indices,
		func(lcv int, chain []*subgraph.Edge) func(b *subgraph.FillableEmbeddingBuilder) bool {
			return func(b *subgraph.FillableEmbeddingBuilder) bool {
				for _, id := range b.Ids {
					if id < 0 {
						continue
					}
					if _, has := seen[id]; !has {
						return false
					}
				}
				return true
			}
	}))
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
			seen[id] = true
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
	if true {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v total-embs %v : %v", len(extensions), len(embeddings), total, pattern.Pretty(dt.G.Colors))
	}
	err = cacheExtsEmbs(dt, pattern, len(embeddings), extensions, embeddings)
	if err != nil {
		return 0, nil, nil, err
	}
	return len(embeddings), extensions, embeddings, nil
}

// unique extensions
func extensions_2(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, error) {
	o := pattern.FindVertexEmbeddings(dt.Indices, dt.Support())
	if o == nil {
		// not supported!
		return nil, cacheExtsEmbs(dt, pattern, 0, nil, nil)
	}

	exts := hashtable.NewLinearHash()
	add := overlapValidExtChecker(dt, func(o *subgraph.Overlap, ext *subgraph.Extension) {
		exts.Put(ext, nil)
	})

	for idx, ids := range o.Ids {
		for x, next := ids.Items()(); next != nil; x, next = next() {
			id := int(x.(types.Int))
			for _, e := range dt.G.Kids[id] {
				add(o, e, idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(o, e, -1, idx)
			}
		}
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Keys()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		extensions = append(extensions, ext)
	}

	return extensions, nil
}

// unique extensions and supported embeddings
func extsAndEmbs_2(dt *Digraph, pattern *subgraph.SubGraph) (int, []*subgraph.Extension, []*subgraph.Embedding, error) {
	if has, sup, exts, embs, err := loadCachedExtsEmbs(dt, pattern); err != nil {
		return 0, nil, nil, err
	} else if has {
		if false {
			errors.Logf("LOAD-DEBUG", "Loaded cached %v exts %v embs %v support %v", pattern, len(exts), len(embs), sup)
		}
		return sup, exts, embs, nil
	}
	// errors.Logf("DEBUG", "----   extsAndEmbs pattern %v", pattern)
	// compute the embeddings
	o := pattern.FindVertexEmbeddings(dt.Indices, dt.Support())
	if o == nil {
		// not supported!
		return 0, nil, nil, cacheExtsEmbs(dt, pattern, 0, nil, nil)
	}

	exts := hashtable.NewLinearHash()
	add := overlapValidExtChecker(dt, func(o *subgraph.Overlap, ext *subgraph.Extension) {
		exts.Put(ext, nil)
	})

	total := 0
	// add the supported embeddings to the vertex sets
	// add the extensions to the extensions set
	// errors.Logf("DEBUG", "parent %v", parent)
	// errors.Logf("DEBUG", "computing embeddings %v", pattern.Pretty(dt.G.Colors))
	// errors.Logf("DEBUG", "computing embeddings %v", pattern)
	for idx, ids := range o.Ids {
		for x, next := ids.Items()(); next != nil; x, next = next() {
			id := int(x.(types.Int))
			for _, e := range dt.G.Kids[id] {
				add(o, e, idx, -1)
			}
			for _, e := range dt.G.Parents[id] {
				add(o, e, -1, idx)
			}
		}
		total++
	}

	// construct the extensions output slice
	extensions := make([]*subgraph.Extension, 0, exts.Size())
	for i, next := exts.Keys()(); next != nil; i, next = next() {
		ext := i.(*subgraph.Extension)
		extensions = append(extensions, ext)
		// errors.Logf("EXT-DEBUG", "o %v ext %v", o, ext)
	}

	// embeddings := o.SupportedEmbeddings(dt.Indices)
	// support := len(embeddings)
	embeddings := []*subgraph.Embedding{}
	support, _ := o.MinSupported()

	// return it all
	if false {
		errors.Logf("CACHE-DEBUG", "Caching exts %v embs %v support %v : %v", len(extensions), len(embeddings), support, o.Pretty(dt.G.Colors))
	}
	err := cacheExtsEmbs(dt, pattern, support, extensions, embeddings)
	if err != nil {
		return 0, nil, nil, err
	}
	return support, extensions, embeddings, nil
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
