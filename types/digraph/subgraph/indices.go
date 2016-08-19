package subgraph

import (
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import ()

type IdColorColor struct {
	Id, EdgeColor, VertexColor int
}

type Colors struct {
	SrcColor, TargColor, EdgeColor int
}


type Indices struct {
	G               *goiso.Graph
	ColorIndex      map[int][]int          // Colors -> []Idx in G.V
	SrcIndex        map[IdColorColor][]int // (SrcIdx, EdgeColor, TargColor) -> TargIdx (where Idx in G.V)
	TargIndex       map[IdColorColor][]int // (TargIdx, EdgeColor, SrcColor) -> SrcIdx (where Idx in G.V)
	EdgeIndex       map[Edge]*goiso.Edge
	EdgeCounts      map[Colors]int         // (src-color, targ-color, edge-color) -> count
	FreqEdges       []Colors               // frequent color triples
	EdgesFromColor  map[int][]Colors       // freq src-colors -> color triples
	EdgesToColor    map[int][]Colors       // freq targ-colors -> color triples
}

func NewIndices(G *goiso.Graph) *Indices {
	return &Indices{
		G: G,
		ColorIndex: make(map[int][]int),
		SrcIndex:   make(map[IdColorColor][]int),
		TargIndex:  make(map[IdColorColor][]int),
		EdgeIndex:  make(map[Edge]*goiso.Edge),
		EdgeCounts: make(map[Colors]int),
		EdgesFromColor: make(map[int][]Colors),
		EdgesToColor: make(map[int][]Colors),
	}
}

func (i *Indices) Colors(e *goiso.Edge) Colors {
	return Colors{
		SrcColor: i.G.V[e.Src].Color,
		TargColor: i.G.V[e.Targ].Color,
		EdgeColor: e.Color,
	}
}

// From an sg.V.id, get the degree of that vertex in the graph.
// so the id is really a Graph Idx
func (i *Indices) Degree(id int) int {
	return len(i.G.Kids[id]) + len(i.G.Parents[id])
}

func (sg *SubGraph) AsIndices(myIndices *Indices, support int) *Indices {
	x := goiso.NewGraph(len(sg.V),len(sg.E))
	g := &x
	for _, color := range myIndices.G.Colors {
		g.AddColor(color)
	}
	for vidx := range sg.V {
		color := g.AddColor(myIndices.G.Colors[sg.V[vidx].Color])
		g.V = append(g.V, goiso.Vertex{Id: vidx, Idx: vidx, Color: color})
	}
	for eidx := range sg.E {
		src := sg.E[eidx].Src
		targ := sg.E[eidx].Targ
		color := g.AddColor(myIndices.G.Colors[sg.E[eidx].Color])
		g.E = append(g.E, goiso.Edge{Arc: goiso.Arc{Src: src, Targ: targ}, Idx: eidx, Color: color})
	}
	indices := NewIndices(g)
	indices.InitVertexIndices()
	indices.InitEdgeIndices(support)
	return indices
}

func intSet(ints []int) types.Set {
	s := set.NewSetMap(hashtable.NewLinearHash())
	for _, i := range ints {
		s.Add(types.Int(i))
	}
	return s
}

func (indices *Indices) InitVertexIndices() {
	for i := range indices.G.V {
		u := &indices.G.V[i]
		indices.ColorIndex[u.Color] = append(indices.ColorIndex[u.Color], u.Idx)
	}
}

func (indices *Indices) InitEdgeIndices(support int) {
	for idx := range indices.G.E {
		e := &indices.G.E[idx]
		edge := Edge{Src: e.Src, Targ: e.Targ, Color: e.Color}
		srcKey := IdColorColor{e.Src, e.Color, indices.G.V[e.Targ].Color}
		targKey := IdColorColor{e.Targ, e.Color, indices.G.V[e.Src].Color}
		colorKey := Colors{indices.G.V[e.Src].Color, indices.G.V[e.Targ].Color, e.Color}
		indices.EdgeIndex[edge] = e
		indices.SrcIndex[srcKey] = append(indices.SrcIndex[srcKey], e.Targ)
		indices.TargIndex[targKey] = append(indices.TargIndex[targKey], e.Src)
		indices.EdgeCounts[colorKey] += 1
	}
	for color, count := range indices.EdgeCounts {
		if count >= support {
			indices.FreqEdges = append(indices.FreqEdges, color)
			indices.EdgesFromColor[color.SrcColor] = append(
				indices.EdgesFromColor[color.SrcColor],
				color)
			indices.EdgesToColor[color.TargColor] = append(
				indices.EdgesToColor[color.TargColor],
				color)
		}
	}
}

func (indices *Indices) IdSet(color int) *set.SortedSet {
	s := set.NewSortedSet(indices.G.ColorFrequency(color))
	for _, gIdx := range indices.ColorIndex[color] {
		s.Add(types.Int(int(gIdx)))
	}
	return s
}

func (indices *Indices) HasEdge(srcId, targId, color int) bool {
	_, has := indices.EdgeIndex[Edge{Src: srcId, Targ: targId, Color: color}]
	return has
}

func (indices *Indices) TargsFromSrc(srcId, edgeColor, targColor int, excludeIds *IdNode, do func(int)) {
outer:
	for _, targId := range indices.SrcIndex[IdColorColor{srcId, edgeColor, targColor}] {
		for c := excludeIds; c != nil; c = c.Prev {
			if targId == c.Id {
				continue outer
			}
		}
		do(targId)
	}
}

func (indices *Indices) SrcsToTarg(targId, edgeColor, srcColor int, excludeIds *IdNode, do func(int)) {
outer:
	for _, srcId := range indices.TargIndex[IdColorColor{targId, edgeColor, srcColor}] {
		for c := excludeIds; c != nil; c = c.Prev {
			if srcId == c.Id {
				continue outer
			}
		}
		do(srcId)
	}
}
