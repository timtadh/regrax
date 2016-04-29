package digraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
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

func extensionPoints(G *goiso.Graph, o *subgraph.Overlap, e *goiso.Edge, src, targ int) []*subgraph.Extension {
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

func validExtChecker(dt *Digraph, do func(*subgraph.Overlap, *subgraph.Extension)) func(*subgraph.Overlap, *goiso.Edge, int, int) {
	return func(o *subgraph.Overlap, e *goiso.Edge, src, targ int) {
		if dt.G.ColorFrequency(e.Color) < dt.Support() {
			return
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < dt.Support() {
			return
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < dt.Support() {
			return
		}
		for _, ep := range extensionPoints(dt.G, o, e, src, targ) {
			if !o.SG.HasExtension(ep) {
				do(o, ep)
			}
		}
	}
}

// unique extensions
func extensions(dt *Digraph, pattern *subgraph.SubGraph) ([]*subgraph.Extension, error) {
	o := pattern.FindVertexEmbeddings(dt.Indices, dt.Support())
	if o == nil {
		// not supported!
		return nil, cacheExtsEmbs(dt, pattern, 0, nil, nil)
	}

	exts := hashtable.NewLinearHash()
	add := validExtChecker(dt, func(o *subgraph.Overlap, ext *subgraph.Extension) {
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
func extsAndEmbs(dt *Digraph, pattern *subgraph.SubGraph) (int, []*subgraph.Extension, []*subgraph.Embedding, error) {
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
	add := validExtChecker(dt, func(o *subgraph.Overlap, ext *subgraph.Extension) {
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
	if true {
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
