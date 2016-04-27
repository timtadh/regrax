package subgraph

import (
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stores/int_int"
)

type IdColorColor struct {
	Id, EdgeColor, VertexColor int
}

type Indices struct {
	G         *goiso.Graph
	ColorMap  int_int.MultiMap
	SrcIndex  map[IdColorColor][]int // (SrcIdx, EdgeColor, TargColor) -> TargIdx (where Idx in G.V)
	TargIndex map[IdColorColor][]int // (TargIdx, EdgeColor, SrcColor) -> SrcIdx (where Idx in G.V)
	EdgeIndex map[Edge]*goiso.Edge
}

func intSet(ints []int) *set.SortedSet {
	s := set.NewSortedSet(len(ints))
	for _, i := range ints {
		s.Add(types.Int(i))
	}
	return s
}

func (indices *Indices) InitColorMap(G *goiso.Graph) error {
	for i := range G.V {
		u := &G.V[i]
		err := indices.ColorMap.Add(int32(u.Color), int32(u.Idx))
		if err != nil {
			return err
		}
	}
	return nil
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

func (indices *Indices) HasEdge(srcId, targId, color int) bool {
	_, has := indices.EdgeIndex[Edge{Src: srcId, Targ: targId, Color: color}]
	return has
}

func (indices *Indices) TargsFromSrc(srcId, edgeColor, targColor int, excludeIds []int) []*goiso.Vertex {
	exclude := intSet(excludeIds)
	targs := make([]*goiso.Vertex, 0, 10)
	for _, targId := range indices.SrcIndex[IdColorColor{srcId, edgeColor, targColor}] {
		if !exclude.Has(types.Int(targId)) {
			targs = append(targs, &indices.G.V[targId])
		}
	}
	return targs
}

func (indices *Indices) SrcsToTarg(targId, edgeColor, srcColor int, excludeIds []int) []*goiso.Vertex {
	exclude := intSet(excludeIds)
	srcs := make([]*goiso.Vertex, 0, 10)
	for _, srcId := range indices.TargIndex[IdColorColor{targId, edgeColor, srcColor}] {
		if !exclude.Has(types.Int(srcId)) {
			srcs = append(srcs, &indices.G.V[srcId])
		}
	}
	return srcs
}
