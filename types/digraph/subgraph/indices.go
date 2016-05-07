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

type Indices struct {
	G          *goiso.Graph
	ColorIndex map[int][]int          // Colors -> []Idx in G.V
	SrcIndex   map[IdColorColor][]int // (SrcIdx, EdgeColor, TargColor) -> TargIdx (where Idx in G.V)
	TargIndex  map[IdColorColor][]int // (TargIdx, EdgeColor, SrcColor) -> SrcIdx (where Idx in G.V)
	EdgeIndex  map[Edge]*goiso.Edge
}

func intSet(ints []int) types.Set {
	s := set.NewSetMap(hashtable.NewLinearHash())
	for _, i := range ints {
		s.Add(types.Int(i))
	}
	return s
}

func (indices *Indices) InitColorMap(G *goiso.Graph) {
	for i := range G.V {
		u := &G.V[i]
		indices.ColorIndex[u.Color] = append(indices.ColorIndex[u.Color], u.Idx)
	}
}

func (indices *Indices) InitEdgeIndices(G *goiso.Graph) {
	for idx := range G.E {
		e := &G.E[idx]
		edge := Edge{Src: e.Src, Targ: e.Targ, Color: e.Color}
		srcKey := IdColorColor{e.Src, e.Color, G.V[e.Targ].Color}
		targKey := IdColorColor{e.Targ, e.Color, G.V[e.Src].Color}
		indices.EdgeIndex[edge] = e
		indices.SrcIndex[srcKey] = append(indices.SrcIndex[srcKey], e.Targ)
		indices.TargIndex[targKey] = append(indices.TargIndex[targKey], e.Src)
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
outer: for _, targId := range indices.SrcIndex[IdColorColor{srcId, edgeColor, targColor}] {
		for c := excludeIds; c != nil; c = c.Prev {
			if targId == c.Id {
				continue outer
			}
		}
		do(targId)
	}
}

func (indices *Indices) SrcsToTarg(targId, edgeColor, srcColor int, excludeIds *IdNode, do func(int)) {
outer: for _, srcId := range indices.TargIndex[IdColorColor{targId, edgeColor, srcColor}] {
		for c := excludeIds; c != nil; c = c.Prev {
			if srcId == c.Id {
				continue outer
			}
		}
		do(srcId)
	}
}
