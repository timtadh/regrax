package digraph

import (
	"math"
	"regexp"
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/pool"
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
	"github.com/timtadh/sfp/stores/subgraph_overlap"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)


type Config struct {
	MinEdges, MaxEdges       int
	MinVertices, MaxVertices int
	Mode                     Mode
	Include, Exclude         *regexp.Regexp
}

type Digraph struct {
	Config
	config                   *config.Config
	G                        *goiso.Graph
	FrequentVertices         [][]byte
	NodeAttrs                int_json.MultiMap
	Embeddings               subgraph_embedding.MultiMap
	Overlap                  subgraph_overlap.MultiMap
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
	pool                     *pool.Pool
	lock                     sync.RWMutex
}

func NewDigraph(config *config.Config, dc *Config) (g *Digraph, err error) {
	if dc.MaxEdges <= 0 {
		dc.MaxEdges = int(math.MaxInt32)
	}
	if dc.MaxVertices <= 0 {
		dc.MaxVertices = int(math.MaxInt32)
	}
	if dc.MinEdges > dc.MaxEdges {
		dc.MinEdges = dc.MaxEdges - 1
	}
	if dc.MinVertices > dc.MaxVertices {
		dc.MinVertices = dc.MaxVertices - 1
	}
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
	var overlap subgraph_overlap.MultiMap = nil
	if dc.Mode&OverlapPruning == OverlapPruning {
		overlap, err = config.SubgraphOverlapMultiMap("digraph-overlap")
		if err != nil {
			return nil, err
		}
	}
	exts, err := config.BytesExtensionMultiMap("digraph-extensions")
	if err != nil {
		return nil, err
	}
	var unexts bytes_extension.MultiMap
	if dc.Mode&ExtensionPruning == ExtensionPruning {
		unexts, err = config.BytesExtensionMultiMap("digraph-unsupported-extensions")
		if err != nil {
			return nil, err
		}
	}
	frequency, err := config.BytesIntMultiMap("digraph-pattern-frequency")
	if err != nil {
		return nil, err
	}
	g = &Digraph{
		Config: *dc,
		NodeAttrs:     nodeAttrs,
		Embeddings:    embeddings,
		Overlap:       overlap,
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
		pool: pool.New(config.Workers()),
	}
	return g, nil
}

func (dt *Digraph) Init(G *goiso.Graph) (err error) {
	dt.lock.Lock()
	dt.G = G
	dt.Indices.G = G
	dt.Indices.InitVertexIndices(G)
	dt.Indices.InitEdgeIndices(G, dt.Support())
	dt.lock.Unlock()

	for color := range dt.Indices.ColorIndex {
		if G.ColorFrequency(color) < dt.config.Support {
			continue
		}
		sg := subgraph.Build(1, 0).FromVertex(color).Build()

		dt.lock.Lock()
		dt.FrequentVertices = append(dt.FrequentVertices, sg.Label())
		dt.lock.Unlock()

		// done for the side effect of saving the Nodes.
		_, _, _, _, err := ExtsAndEmbs(dt, sg, nil, nil, dt.Mode, false)
		if err != nil {
			return err
		}
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
	return NewEmbListNode(g, subgraph.EmptySubGraph(), nil, nil, nil)
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
	g.lock.Lock()
	defer g.lock.Unlock()
	g.config.AsyncTasks.Wait()
	g.Parents.Close()
	g.ParentCount.Close()
	g.Children.Close()
	g.ChildCount.Close()
	g.CanonKids.Close()
	g.CanonKidCount.Close()
	g.Embeddings.Close()
	if g.Overlap != nil {
		g.Overlap.Close()
	}
	g.Extensions.Close()
	if g.UnsupExts != nil {
		g.UnsupExts.Close()
	}
	g.NodeAttrs.Close()
	g.Frequency.Close()
	return nil
}
