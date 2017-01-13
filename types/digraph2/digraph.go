package digraph2

import (
	"math"
	"regexp"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph2/digraph"
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)


type Config struct {
	MinEdges, MaxEdges       int
	MinVertices, MaxVertices int
	Include, Exclude         *regexp.Regexp
	EmbSearchStartPoint      subgraph.EmbSearchStartPoint
}

type Digraph struct {
	Config
	config                   *config.Config
	G                        *digraph.Digraph
	Labels                   *digraph.Labels
	FrequentVertices         []*Node
	Indices                  *digraph.Indices
	NodeAttrs                map[int]map[string]interface{}
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
	g = &Digraph{
		Config: *dc,
		config: config,
	}
	return g, nil
}

func (dt *Digraph) Init(b *digraph.Builder, l *digraph.Labels) (err error) {
	i := digraph.NewIndices(b, dt.config.Support)
	errors.Logf("DEBUG", "done building indices")
	dt.G = i.G
	dt.Indices = i
	dt.Labels = l

	errors.Logf("DEBUG", "computing starting points")
	// for color, _ := range dt.Indices.ColorIndex {
	// 	// sg := subgraph.Build(1, 0).FromVertex(color).Build()
	// 	// dt.FrequentVertices = append(dt.FrequentVertices, n)
	// }
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

func (g *Digraph) Root() lattice.Node {
	// return RootEmbListNode(g)
	return nil
}

func VE(node lattice.Node) (V, E int) {
	panic("wizard")
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
	return nil
}
