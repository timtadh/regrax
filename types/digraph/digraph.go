package digraph

import ()

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_extension"
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/stores/int_json"
	"github.com/timtadh/sfp/stores/subgraph_embedding"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

type Digraph struct {
	MinEdges, MaxEdges       int
	MinVertices, MaxVertices int
	Mode                     Mode
	G                        *goiso.Graph
	FrequentVertices         [][]byte
	NodeAttrs                int_json.MultiMap
	Embeddings               subgraph_embedding.MultiMap
	Extensions               bytes_extension.MultiMap
	UnsupExts                bytes_extension.MultiMap
	Parents                  bytes_bytes.MultiMap
	ParentCount              bytes_int.MultiMap
	Children                 bytes_bytes.MultiMap
	ChildCount               bytes_int.MultiMap
	CanonKids                bytes_bytes.MultiMap
	CanonKidCount            bytes_int.MultiMap
	Frequency                bytes_int.MultiMap
	Indices                  *subgraph.Indices
	config                   *config.Config
}

func NewDigraph(config *config.Config, mode Mode, minE, maxE, minV, maxV int) (g *Digraph, err error) {
	nodeAttrs, err := config.IntJsonMultiMap("digraph-node-attrs")
	if err != nil {
		return nil, err
	}
	parents, err := config.MultiMap("digraph-parents")
	if err != nil {
		return nil, err
	}
	parentCount, err := config.BytesIntMultiMap("digraph-parent-count")
	if err != nil {
		return nil, err
	}
	children, err := config.MultiMap("digraph-children")
	if err != nil {
		return nil, err
	}
	childCount, err := config.BytesIntMultiMap("digraph-child-count")
	if err != nil {
		return nil, err
	}
	canonKids, err := config.MultiMap("digraph-canon-kids")
	if err != nil {
		return nil, err
	}
	canonKidCount, err := config.BytesIntMultiMap("digraph-canon-kid-count")
	if err != nil {
		return nil, err
	}
	embeddings, err := config.SubgraphEmbeddingMultiMap("digraph-embeddings")
	if err != nil {
		return nil, err
	}
	exts, err := config.BytesExtensionMultiMap("digraph-extensions")
	if err != nil {
		return nil, err
	}
	unexts, err := config.BytesExtensionMultiMap("digraph-unsupported-extensions")
	if err != nil {
		return nil, err
	}
	frequency, err := config.BytesIntMultiMap("digraph-pattern-frequency")
	if err != nil {
		return nil, err
	}
	g = &Digraph{
		MinEdges:      minE,
		MaxEdges:      maxE,
		MinVertices:   minV,
		MaxVertices:   maxV,
		Mode:          mode,
		NodeAttrs:     nodeAttrs,
		Embeddings:    embeddings,
		Extensions:    exts,
		UnsupExts:     unexts,
		Parents:       parents,
		ParentCount:   parentCount,
		Children:      children,
		ChildCount:    childCount,
		CanonKids:     canonKids,
		CanonKidCount: canonKidCount,
		Frequency:     frequency,
		Indices: &subgraph.Indices{
			ColorIndex: make(map[int][]int),
			SrcIndex:   make(map[subgraph.IdColorColor][]int),
			TargIndex:  make(map[subgraph.IdColorColor][]int),
			EdgeIndex:  make(map[subgraph.Edge]*goiso.Edge),
			EdgeCounts: make(map[subgraph.Colors]int),
		},
		config: config,
	}
	return g, nil
}

func (dt *Digraph) Init(G *goiso.Graph) (err error) {
	dt.G = G
	dt.Indices.G = G

	for i := range G.V {
		u := &G.V[i]
		dt.Indices.ColorIndex[u.Color] = append(dt.Indices.ColorIndex[u.Color], u.Idx)
		if G.ColorFrequency(u.Color) >= dt.config.Support {
			emb := subgraph.BuildEmbedding(1, 0).FromVertex(u.Color, u.Idx).Build()
			err := dt.Embeddings.Add(emb.SG, emb)
			if err != nil {
				return err
			}
		}
	}

	dt.Indices.InitEdgeIndices(G)

	err = subgraph_embedding.DoKey(dt.Embeddings.Keys, func(sg *subgraph.SubGraph) error {
		dt.FrequentVertices = append(dt.FrequentVertices, sg.Label())
		exts, err := extensions(dt, sg)
		if err != nil {
			return err
		}
		color := sg.V[0].Color
		err = dt.Frequency.Add(sg.Label(), int32(G.ColorFrequency(color)))
		if err != nil {
			return err
		}
		for _, ext := range exts {
			err := dt.Extensions.Add(sg.Label(), ext)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Digraph) Support() int {
	return g.config.Support
}

func (g *Digraph) LargestLevel() int {
	return g.MaxEdges
}

func (g *Digraph) MinimumLevel() int {
	if g.MinEdges > 0 {
		return g.MinEdges
	} else if g.MinVertices > 0 {
		return g.MinVertices - 1
	}
	return 0
}

func RootEmbListNode(g *Digraph) *EmbListNode {
	return NewEmbListNode(g, subgraph.EmptySubGraph(), nil, nil)
}

func (g *Digraph) Root() lattice.Node {
	return RootEmbListNode(g)
}

func VE(node lattice.Node) (V, E int) {
	E = 0
	V = 0
	switch n := node.(type) {
	case *EmbListNode:
		return len(n.Pat.V), len(n.Pat.E)
	default:
		panic(errors.Errorf("unknown node type %T %v", node, node))
	}
	return V, E
}

func (g *Digraph) Acceptable(node lattice.Node) bool {
	V, E := VE(node)
	return g.MinEdges <= E && E <= g.MaxEdges && g.MinVertices <= V && V <= g.MaxVertices
}

func (g *Digraph) TooLarge(node lattice.Node) bool {
	V, E := VE(node)
	return E > g.MaxEdges || V > g.MaxVertices
}

func (g *Digraph) Close() error {
	g.config.AsyncTasks.Wait()
	g.Parents.Close()
	g.ParentCount.Close()
	g.Children.Close()
	g.ChildCount.Close()
	g.CanonKids.Close()
	g.CanonKidCount.Close()
	g.Embeddings.Close()
	g.Extensions.Close()
	g.UnsupExts.Close()
	g.NodeAttrs.Close()
	g.Frequency.Close()
	return nil
}
