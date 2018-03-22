package digraph

import (
	"math"
	"regexp"
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/pool"
)

import (
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/stores/int_json"
	"github.com/timtadh/regrax/types/digraph/digraph"
	"github.com/timtadh/regrax/types/digraph/subgraph"
)

type Config struct {
	MinEdges, MaxEdges       int
	MinVertices, MaxVertices int
	Mode                     Mode
	Include, Exclude         *regexp.Regexp
	EmbSearchStartPoint      subgraph.EmbSearchStartPoint
}

type Digraph struct {
	Config
	config           *config.Config
	G                *digraph.Digraph
	Labels           *digraph.Labels
	FrequentVertices []*EmbListNode
	NodeAttrs        int_json.MultiMap
	Indices          *digraph.Indices
	pool             *pool.Pool
	lock             sync.RWMutex
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
	g = &Digraph{
		Config:    *dc,
		NodeAttrs: nodeAttrs,
		config:    config,
		pool:      pool.New(config.Workers()),
	}
	return g, nil
}

func (dt *Digraph) Init(b *digraph.Builder, l *digraph.Labels) (err error) {
	dt.lock.Lock()
	// i := digraph.NewIndices(b, dt.config.Support, dt.Mode & ExtFromFreqEdges == ExtFromFreqEdges)
	i := digraph.NewIndices(b, dt.config.Support)
	errors.Logf("DEBUG", "done building indices")
	dt.G = i.G
	dt.Indices = i
	dt.Labels = l
	dt.lock.Unlock()

	errors.Logf("DEBUG", "computing starting points")
	for color, _ := range dt.Indices.ColorIndex {
		sg := subgraph.Build(1, 0).FromVertex(color).Build()
		_, exts, embs, _, err := ExtsAndEmbs(dt, sg, nil, nil, dt.Mode, false)
		if err != nil {
			return err
		}
		n := NewEmbListNode(dt, sg, exts, embs, nil)
		dt.lock.Lock()
		dt.FrequentVertices = append(dt.FrequentVertices, n)
		dt.lock.Unlock()
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
	g.NodeAttrs.Close()
	return nil
}
