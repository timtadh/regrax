package subgraph

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

type Overlap struct {
	SG  *SubGraph
	Ids []*set.SortedSet // the embeddings for each vertex
}

func (sg *SubGraph) FindVertexEmbeddings(G *goiso.Graph, indices *Indices, minSupport int) (*Overlap, error) {
	embs := make([]*set.SortedSet, 0, len(sg.V))
	for range sg.V {
		embs = append(embs, set.NewSortedSet(minSupport))
	}
	return sg.findOverlap(G, indices, minSupport)
}

func (sg *SubGraph) findOverlap(G *goiso.Graph, indices *Indices, minSupport int) (*Overlap, error) {
	chain := sg.edgeChain()
	v0, err := sg.idSet(indices, sg.V[0].Color)
	if err != nil {
		return nil, err
	}
	b := BuildOverlap(len(sg.V), len(sg.E)).Fillable().Ctx(func(b *FillableOverlapBuilder) {
		b.SetVertex(0, sg.V[0].Color, v0)
	})
	for _, e := range chain {
		errors.Logf("VE-DEBUG", "edge %v", e)
		// o.addEdge(e)
	}
	errors.Logf("VE-DEBUG", "so far %v", b)
	return nil, errors.Errorf("unfinished")
}

func (sg *SubGraph) idSet(indices *Indices, color int) (*set.SortedSet, error) {
	s := set.NewSortedSet(10)
	err := indices.ColorMap.DoFind(int32(color), func(color, gIdx int32) error {
		s.Add(types.Int(int(gIdx)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (o *Overlap) String() string {
	V := make([]string, 0, len(o.SG.V))
	E := make([]string, 0, len(o.SG.E))
	for i := range o.SG.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			o.SG.V[i].Color,
			o.Ids[i],
		))
	}
	for _, e := range o.SG.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			e.Color,
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(o.SG.E), len(o.SG.V), strings.Join(V, ""), strings.Join(E, ""))
}
