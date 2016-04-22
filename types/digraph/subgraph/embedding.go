package subgraph

import (
	"encoding/binary"
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
)


type Embedding struct {
	SG *SubGraph
	Ids []int // Idx vertices in the graph
}



func (emb *Embedding) MarshalBinary() ([]byte, error) {
	return emb.Label(), nil
}

func (emb *Embedding) UnmarshalBinary(bytes []byte) error {
	if emb.SG != nil || emb.Ids != nil {
		return errors.Errorf("Embedding is already loaded! will not load serialized data")
	}
	if len(bytes) < 8 {
		return errors.Errorf("bytes was too small %v < 8", len(bytes))
	}
	lenE := int(binary.BigEndian.Uint32(bytes[0:4]))
	lenV := int(binary.BigEndian.Uint32(bytes[4:8]))
	off := 8
	expected := 8 + lenV*8 + lenE*12
	if len(bytes) < expected {
		return errors.Errorf("bytes was too small %v < %v", len(bytes), expected)
	}
	ids := make([]int, lenV)
	sg := &SubGraph{
		V: make([]Vertex, lenV),
		E: make([]Edge, lenE),
		Adj: make([][]int, lenV),
	}
	for i := 0; i < lenV; i++ {
		s := off + i*8
		e := s + 4
		id := int(binary.BigEndian.Uint32(bytes[s:e]))
		s += 4
		e += 4
		color := int(binary.BigEndian.Uint32(bytes[s:e]))
		ids[i] = id
		sg.V[i].Idx = i
		sg.V[i].Color = color
		sg.Adj[i] = make([]int, 0, 5)
	}
	off += lenV * 8
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
	emb.SG = sg
	emb.Ids = ids
	return nil
}

func (emb *Embedding) Label() []byte {
	size := 8 + len(emb.SG.V)*8 + len(emb.SG.E)*12
	label := make([]byte, size)
	binary.BigEndian.PutUint32(label[0:4], uint32(len(emb.SG.E)))
	binary.BigEndian.PutUint32(label[4:8], uint32(len(emb.SG.V)))
	off := 8
	for i := range emb.SG.V {
		s := off + i*8
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(emb.SG.V[i].Color))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(emb.Ids[i]))
	}
	off += len(emb.SG.V) * 8
	for i := range emb.SG.E {
		edge := &emb.SG.E[i]
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

func (emb *Embedding) String() string {
	V := make([]string, 0, len(emb.SG.V))
	E := make([]string, 0, len(emb.SG.E))
	for i := range emb.SG.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			emb.Ids[i],
			emb.SG.V[i].Color,
		))
	}
	for i := range emb.SG.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			emb.SG.E[i].Src,
			emb.SG.E[i].Targ,
			emb.SG.E[i].Color,
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(emb.SG.E), len(emb.SG.V), strings.Join(V, ""), strings.Join(E, ""))
}

