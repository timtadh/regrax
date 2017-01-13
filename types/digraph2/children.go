package digraph2

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
	builder := n.SubGraph.Builder()
	for ext, embs := range n.extensions() {
		support := n.support(embs)
		if support < n.dt.Support() {
			continue
		}
		b := builder.Copy()
		_, _, err := b.Extend(&ext)
		if err != nil {
			return nil, err
		}
		extended := b.Build()
		if allow != nil {
			allowed, err := allow(extended)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		nodes = append(nodes, NewNode(n.dt, extended, embs))
	}
	return nodes, nil
}

func (n *Node) extensions() map[subgraph.Extension][]*subgraph.Embedding {
	exts := make(map[subgraph.Extension][]*subgraph.Embedding)
	add := n.validExtChecker(func(emb *subgraph.Embedding, ext *subgraph.Extension) {
		exts[*ext] = append(exts[*ext], emb)
	})
	for _, embedding := range n.Embeddings {
		for emb := embedding; emb != nil; emb = emb.Prev {
			for _, e := range n.dt.G.Kids[emb.EmbIdx] {
				edge := &n.dt.G.E[e]
				add(emb, edge, emb.SgIdx, -1, edge.Src)
			}
			for _, e := range n.dt.G.Parents[emb.EmbIdx] {
				edge := &n.dt.G.E[e]
				add(emb, edge, -1, emb.SgIdx, edge.Targ)
			}
		}
	}
	return exts
}

func (n *Node) validExtChecker(do func(*subgraph.Embedding, *subgraph.Extension)) func (*subgraph.Embedding, *digraph.Edge, int, int, int) {
	return func(emb *subgraph.Embedding, e *digraph.Edge, src, targ, embIdx int) {
		if n.dt.Indices.EdgeCounts[n.dt.Indices.Colors(e)] < n.dt.Support() {
			return
		}
		emb, ext := n.extension(emb, e, src, targ, embIdx)
		if n.SubGraph.HasExtension(ext) {
			errors.Logf("DEBUG", "had extension")
			errors.Logf("DEBUG", "sg %v", n)
			errors.Logf("DEBUG", "ext %v", ext)
			return
		} else {
			errors.Logf("DEBUG", "did not have extension")
			errors.Logf("DEBUG", "sg %v", n)
			errors.Logf("DEBUG", "ext %v", ext)
			do(emb, ext)
		}
	}
}

func (n *Node) extension(embedding *subgraph.Embedding, e *digraph.Edge, src, targ, embIdx int) (*subgraph.Embedding, *subgraph.Extension) {
	hasTarg := false
	hasSrc := false
	var srcIdx int = len(n.SubGraph.V)
	var targIdx int = len(n.SubGraph.V)
	var newVE *subgraph.VertexEmbedding = nil
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
	ext := subgraph.NewExt(
		subgraph.Vertex{Idx: srcIdx, Color: n.dt.G.V[e.Src].Color},
		subgraph.Vertex{Idx: targIdx, Color: n.dt.G.V[e.Targ].Color},
		e.Color)
	if newVE != nil {
		return embedding.Extend(*newVE), ext
	}
	return embedding, ext
}
