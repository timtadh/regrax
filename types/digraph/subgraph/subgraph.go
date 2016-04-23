package subgraph

import (
	"encoding/binary"
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
	"github.com/timtadh/goiso/bliss"
)

import ()

type SubGraph struct {
	V   Vertices
	E   Edges
	Adj [][]int
}

type Vertices []Vertex
type Edges []Edge

type Vertex struct {
	Idx   int
	Color int
}

type Edge struct {
	Src, Targ, Color int
}


func EmptySubGraph() *SubGraph {
	return &SubGraph{}
}

func (V Vertices) Iterate() (vi bliss.VertexIterator) {
	i := 0
	vi = func() (color int, _ bliss.VertexIterator) {
		if i >= len(V) {
			return 0, nil
		}
		color = V[i].Color
		i++
		return color, vi
	}
	return vi
}

func (E Edges) Iterate() (ei bliss.EdgeIterator) {
	i := 0
	ei = func() (src, targ, color int, _ bliss.EdgeIterator) {
		if i >= len(E) {
			return 0, 0, 0, nil
		}
		src = E[i].Src
		targ = E[i].Targ
		color = E[i].Color
		i++
		return src, targ, color, ei
	}
	return ei
}

// Note since *SubGraphs are constructed from *goiso.SubGraphs they are in
// canonical ordering. This is a necessary assumption for Embeddings() to
// work properly.
func FromEmbedding(sg *goiso.SubGraph) *SubGraph {
	if sg == nil {
		return &SubGraph{
			V:   make([]Vertex, 0),
			E:   make([]Edge, 0),
			Adj: make([][]int, 0),
		}
	}
	pat := &SubGraph{
		V:   make([]Vertex, len(sg.V)),
		E:   make([]Edge, len(sg.E)),
		Adj: make([][]int, len(sg.V)),
	}
	for i := range sg.V {
		pat.V[i].Idx = i
		pat.V[i].Color = sg.V[i].Color
		pat.Adj[i] = make([]int, 0, 5)
	}
	for i := range sg.E {
		pat.E[i].Src = sg.E[i].Src
		pat.E[i].Targ = sg.E[i].Targ
		pat.E[i].Color = sg.E[i].Color
		pat.Adj[pat.E[i].Src] = append(pat.Adj[pat.E[i].Src], i)
		pat.Adj[pat.E[i].Targ] = append(pat.Adj[pat.E[i].Targ], i)
	}
	return pat
}

func FromLabel(label []byte) (*SubGraph, error) {
	sg := new(SubGraph)
	err := sg.UnmarshalBinary(label)
	if err != nil {
		return nil, err
	}
	return sg, nil
}

func (sg *SubGraph) Builder() *Builder {
	return Build(len(sg.V), len(sg.E)).From(sg)
}

func (sg *SubGraph) MarshalBinary() ([]byte, error) {
	return sg.Label(), nil
}

func (sg *SubGraph) UnmarshalBinary(bytes []byte) error {
	if sg.V != nil || sg.E != nil || sg.Adj != nil {
		return errors.Errorf("sg is already loaded! will not load serialized data")
	}
	if len(bytes) < 8 {
		return errors.Errorf("bytes was too small %v < 8", len(bytes))
	}
	lenE := int(binary.BigEndian.Uint32(bytes[0:4]))
	lenV := int(binary.BigEndian.Uint32(bytes[4:8]))
	off := 8
	expected := 8 + lenV*4 + lenE*12
	if len(bytes) < expected {
		return errors.Errorf("bytes was too small %v < %v", len(bytes), expected)
	}
	sg.V = make([]Vertex, lenV)
	sg.E = make([]Edge, lenE)
	sg.Adj = make([][]int, lenV)
	for i := 0; i < lenV; i++ {
		s := off + i*4
		e := s + 4
		color := int(binary.BigEndian.Uint32(bytes[s:e]))
		sg.V[i].Idx = i
		sg.V[i].Color = color
		sg.Adj[i] = make([]int, 0, 5)
	}
	off += lenV * 4
	for i := 0; i < lenE; i++ {
		s := off + i*12
		e := s + 4
		src := int(binary.BigEndian.Uint32(bytes[s:e]))
		s += 4
		e += 4
		targ := int(binary.BigEndian.Uint32(bytes[s:e]))
		s += 4
		e += 4
		color := int(binary.BigEndian.Uint32(bytes[s:e]))
		sg.E[i].Src = src
		sg.E[i].Targ = targ
		sg.E[i].Color = color
		sg.Adj[src] = append(sg.Adj[src], i)
		sg.Adj[targ] = append(sg.Adj[targ], i)
	}
	return nil
}

func (sg *SubGraph) Label() []byte {
	size := 8 + len(sg.V)*4 + len(sg.E)*12
	label := make([]byte, size)
	binary.BigEndian.PutUint32(label[0:4], uint32(len(sg.E)))
	binary.BigEndian.PutUint32(label[4:8], uint32(len(sg.V)))
	off := 8
	for i, v := range sg.V {
		s := off + i*4
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(v.Color))
	}
	off += len(sg.V) * 4
	for i, edge := range sg.E {
		s := off + i*12
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Src))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Targ))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(edge.Color))
	}
	return label
}

func (sg *SubGraph) String() string {
	V := make([]string, 0, len(sg.V))
	E := make([]string, 0, len(sg.E))
	for _, v := range sg.V {
		V = append(V, fmt.Sprintf(
			"(%v)",
			v.Color,
		))
	}
	for _, e := range sg.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			e.Src,
			e.Targ,
			e.Color,
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(sg.E), len(sg.V), strings.Join(V, ""), strings.Join(E, ""))
}

