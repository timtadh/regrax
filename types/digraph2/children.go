package digraph2

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph2/digraph"
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)

func (n *Node) findChildren(allow func(*subgraph.SubGraph) (bool, error)) (nodes []lattice.Node, err error) {
	if false {
		errors.Logf("DEBUG", "findChildren %v", n)
	}
	if n.SubGraph == nil {
		for _, n := range n.dt.FrequentVertices { 
			nodes = append(nodes, n)
		}
		return nodes, nil
	}
	unsupported := n.unsupportedExts
	vords := make([][]int, 0, 10)
	builder := n.SubGraph.Builder()
	seen := make(map[string]subgraph.Extension)
	exts := n.extensions(unsupported)
	for ext, embs := range exts {
		// ext := p.ext
		// embs := p.embs
		if len(embs) < n.dt.Support() {
			unsupported[ext] = true
			continue
		}
		b := builder.Copy()
		_, _, err := b.Extend(&ext)
		if err != nil {
			return nil, err
		}
		support := n.support(len(b.V), embs)
		if support < n.dt.Support() {
			unsupported[ext] = true
			continue
		}
		vord, eord := b.CanonicalPermutation()
		extended := b.BuildFromPermutation(vord, eord)
		if allow != nil {
			allowed, err := allow(extended)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		tembs := embs.Translate(len(extended.V), vord)
		// for _, temb := range tembs {
		// 	if !extended.EmbeddingExists(temb, n.dt.G) {
		// 		panic("non-existant embedding")
		// 	}
		// }
		label := string(extended.Label())
		if _, has := seen[label]; has {
			// errors.Logf("ERROR", "sg %v", n.SubGraph)
			// errors.Logf("ERROR", "cur sg %v", extended)
			// errors.Logf("ERROR", "cur ext %v", ext)
			// prevExt := seen[label]
			// b := builder.Copy()
			// b.Extend(&prevExt)
			// prevSg := b.Build()
			// errors.Logf("ERROR", "prev sg %v", prevSg)
			// errors.Logf("ERROR", "prev ext %v", prevExt)
			// m := make(map[subgraph.Extension][]*subgraph.SubGraph)
			// m[prevExt] = append(m[prevExt], prevSg)
			// m[ext] = append(m[ext], extended)
			// errors.Logf("ERROR", "map %v", m)
			// for ext := range exts {
			// 	errors.Logf("ERROR", "ext %v", ext)
			// }
			// panic("seen node before")
			continue
		}
		seen[label] = ext
		nodes = append(nodes, NewNode(n.dt, extended, tembs))
		vords = append(vords, vord)
	}
	for i, c := range nodes {
		c.(*Node).addUnsupportedExts(unsupported, len(n.SubGraph.V), vords[i])
	}
	return nodes, nil
}

func (n *Node) extensions(unsupported map[subgraph.Extension]bool) map[subgraph.Extension]subgraph.Embeddings {
	// exts := make(extensions, 0, 10)
	exts := make(map[subgraph.Extension]subgraph.Embeddings)
	add := n.validExtChecker(unsupported, func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		// exts = append(exts, extension{ext, emb})
		exts[*ext] = append(exts[*ext], emb)
	})
	for _, embedding := range n.Embeddings {
		for emb := embedding; emb != nil; emb = emb.Prev {
			for _, e := range n.dt.G.Kids[emb.EmbIdx] {
				edge := &n.dt.G.E[e]
				add(embedding, edge, emb.SgIdx, -1, edge.Targ)
			}
			for _, e := range n.dt.G.Parents[emb.EmbIdx] {
				edge := &n.dt.G.E[e]
				add(embedding, edge, -1, emb.SgIdx, edge.Src)
			}
		}
	}
	// return exts.partition()
	return exts
}

func (n *Node) validExtChecker(unsupported map[subgraph.Extension]bool, do func(*subgraph.Embedding, *subgraph.Extension)) func (*subgraph.Embedding, *digraph.Edge, int, int, int) {
	return func(emb *subgraph.Embedding, e *digraph.Edge, src, targ, embIdx int) {
		if n.dt.Indices.EdgeCounts[n.dt.Indices.Colors(e)] < n.dt.Support() {
			return
		}
		emb, ext := n.extension(emb, e, src, targ, embIdx)
		if n.SubGraph.HasExtension(ext) {
			return
		}
		if unsupported[*ext] {
			return
		}
		do(emb, ext)
	}
}

func (n *Node) extension(embedding *subgraph.Embedding, e *digraph.Edge, src, targ, embIdx int) (*subgraph.Embedding, *subgraph.Extension) {
	hasTarg := false
	hasSrc := false
	var srcIdx int = len(n.SubGraph.V)
	var targIdx int = len(n.SubGraph.V)
	if src >= 0 {
		hasSrc = true
		srcIdx = src
	}
	if targ >= 0 {
		hasTarg = true
		targIdx = targ
	}
	for emb := embedding; emb != nil; emb = emb.Prev {
		if hasTarg && hasSrc {
			break
		}
		if !hasSrc && e.Src == emb.EmbIdx {
			hasSrc = true
			srcIdx = emb.SgIdx
		}
		if !hasTarg && e.Targ == emb.EmbIdx {
			hasTarg = true
			targIdx = emb.SgIdx
		}
	}
	ext := subgraph.NewExt(
		subgraph.Vertex{Idx: srcIdx, Color: n.dt.G.V[e.Src].Color},
		subgraph.Vertex{Idx: targIdx, Color: n.dt.G.V[e.Targ].Color},
		e.Color)
	var newVE *subgraph.VertexEmbedding = nil
	if !hasSrc && !hasTarg {
		panic("both src and targ unattached")
	} else if !hasSrc {
		newVE = &subgraph.VertexEmbedding{
			SgIdx: srcIdx,
			EmbIdx: embIdx,
		}
	} else if !hasTarg {
		newVE = &subgraph.VertexEmbedding{
			SgIdx: targIdx,
			EmbIdx: embIdx,
		}
	}
	if newVE != nil {
		embedding = embedding.Extend(*newVE)
	}
	return embedding, ext
}
