package digraph

import (
	"runtime"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/stores/bytes_subgraph"
	"github.com/timtadh/sfp/stores/bytes_extension"
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/stores/int_json"
	"github.com/timtadh/sfp/types/digraph/ext"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)



type Digraph struct {
	MinEdges, MaxEdges       int
	MinVertices, MaxVertices int
	G                        *goiso.Graph
	FrequentVertices         [][]byte
	Supported                Supported
	Extender                 *ext.Extender
	NodeAttrs                int_json.MultiMap
	Embeddings               bytes_subgraph.MultiMap
	Extensions               bytes_extension.MultiMap
	Parents                  bytes_bytes.MultiMap
	ParentCount              bytes_int.MultiMap
	Children                 bytes_bytes.MultiMap
	ChildCount               bytes_int.MultiMap
	CanonKids                bytes_bytes.MultiMap
	CanonKidCount            bytes_int.MultiMap
	ColorMap                 int_int.MultiMap
	Frequency                bytes_int.MultiMap
	config                   *config.Config
}

func NewDigraph(config *config.Config, sup Supported, minE, maxE, minV, maxV int) (g *Digraph, err error) {
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
	colorMap, err := config.IntIntMultiMap("digraph-color-map")
	if err != nil {
		return nil, err
	}
	exts, err := config.BytesExtensionMultiMap("digraph-extensions")
	if err != nil {
		return nil, err
	}
	frequency, err := config.BytesIntMultiMap("digraph-pattern-frequency")
	if err != nil {
		return nil, err
	}
	g = &Digraph{
		Supported:     sup,
		Extender:      ext.NewExtender(runtime.NumCPU()),
		MinEdges:      minE,
		MaxEdges:      maxE,
		MinVertices:   minV,
		MaxVertices:   maxV,
		NodeAttrs:     nodeAttrs,
		Extensions:    exts,
		Parents:       parents,
		ParentCount:   parentCount,
		Children:      children,
		ChildCount:    childCount,
		CanonKids:     canonKids,
		CanonKidCount: canonKidCount,
		ColorMap:      colorMap,
		Frequency:     frequency,
		config:        config,
	}
	return g, nil
}

func (dt *Digraph) Init(G *goiso.Graph) (err error) {
	dt.G = G
	dt.Embeddings, err = dt.config.BytesSubgraphMultiMap("digraph-embeddings", bytes_subgraph.DeserializeSubGraph(G))
	if err != nil {
		return err
	}

	for i := range G.V {
		u := &G.V[i]
		err = dt.ColorMap.Add(int32(u.Color), int32(u.Idx))
		if err != nil {
			return err
		}
		if G.ColorFrequency(u.Color) >= dt.config.Support {
			sg, _ := G.VertexSubGraph(u.Idx)
			err := dt.Embeddings.Add(sg.ShortLabel(), sg)
			if err != nil {
				return err
			}
		}
	}

	err = bytes_subgraph.DoKey(dt.Embeddings.Keys, func(label []byte) error {
		dt.FrequentVertices = append(dt.FrequentVertices, label)
		pat, err := subgraph.FromLabel(label)
		if err != nil {
			return err
		}
		exts, err := extensions(dt, pat)
		if err != nil {
			return err
		}
		color := pat.V[0].Color
		err = dt.Frequency.Add(label, int32(G.ColorFrequency(color)))
		if err != nil {
			return err
		}
		for _, ext := range exts {
			err := dt.Extensions.Add(label, ext)
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
	return NewEmbListNode(g, nil, nil)
}

func (g *Digraph) Root() lattice.Node {
	return RootEmbListNode(g)
}

func VE(node lattice.Node) (V, E int) {
	E = 0
	V = 0
	switch n := node.(type) {
	case *EmbListNode:
		if len(n.embeddings) > 0 {
			E = len(n.embeddings[0].E)
			V = len(n.embeddings[0].V)
		}
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
	g.Extender.Stop()
	g.Parents.Close()
	g.ParentCount.Close()
	g.Children.Close()
	g.ChildCount.Close()
	g.CanonKids.Close()
	g.CanonKidCount.Close()
	g.Embeddings.Close()
	g.Extensions.Close()
	g.NodeAttrs.Close()
	g.ColorMap.Close()
	g.Frequency.Close()
	return nil
}

