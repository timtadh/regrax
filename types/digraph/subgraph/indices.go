package subgraph


import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stores/int_int"
)

type IndexKey struct {
	Id, EdgeColor, VertexColor int
}

type Indices struct {
	ColorMap  int_int.MultiMap
	SrcIndex  map[IndexKey][]int // (SrcIdx, EdgeColor, TargColor) -> TargIdx (where Idx in G.V)
	TargIndex map[IndexKey][]int // (TargIdx, EdgeColor, SrcColor) -> SrcIdx (where Idx in G.V)
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
		srcKey := IndexKey{e.Src, e.Color, G.V[e.Targ].Color}
		targKey := IndexKey{e.Targ, e.Color, G.V[e.Src].Color}
		if _, has := indices.SrcIndex[srcKey]; !has {
			indices.SrcIndex[srcKey] = make([]int, 0, 1)
		}
		if _, has := indices.TargIndex[targKey]; !has {
			indices.TargIndex[targKey] = make([]int, 0, 1)
		}
		indices.SrcIndex[srcKey] = append(indices.SrcIndex[srcKey], e.Targ)
		indices.TargIndex[targKey] = append(indices.TargIndex[targKey], e.Src)
	}
}

