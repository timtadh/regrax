package subgraph

import (
	"encoding/binary"
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)


type Embedding struct {
	SG *SubGraph
	Ids []int // Idx vertices in the graph
}


func LoadEmbedding(bytes []byte) (*Embedding, error) {
	emb := new(Embedding)
	err := emb.UnmarshalBinary(bytes)
	if err != nil {
		return nil, err
	}
	return emb, nil
}

func (emb *Embedding) Builder() *EmbeddingBuilder {
	return BuildEmbedding(len(emb.SG.V), len(emb.SG.E)).From(emb)
}

func (emb *Embedding) HasExtension(ext *Extension) bool {
	if ext.Source.Idx >= len(emb.SG.V) || ext.Source.Color != emb.SG.V[ext.Source.Idx].Color {
		return false
	}
	if ext.Target.Idx >= len(emb.SG.V) || ext.Target.Color != emb.SG.V[ext.Target.Idx].Color {
		return false
	}
	for _, eidx := range emb.SG.Adj[ext.Source.Idx] {
		e := &emb.SG.E[eidx]
		if e.Src == ext.Source.Idx && e.Targ == ext.Target.Idx && e.Color == ext.Color {
			return true
		}
	}
	return false
}

func (emb *Embedding) Exists(G *goiso.Graph) bool {
	for i := range emb.SG.E {
		e := &emb.SG.E[i]
		found := false
		for _, ke := range G.Kids[emb.Ids[e.Src]] {
			if ke.Color != e.Color {
				continue
			}
			if G.V[ke.Src].Color != emb.SG.V[e.Src].Color {
				continue
			}
			if G.V[ke.Targ].Color != emb.SG.V[e.Targ].Color {
				continue
			}
			if ke.Src != emb.Ids[e.Src] {
				continue
			}
			if ke.Targ != emb.Ids[e.Targ] {
				continue
			}
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func (emb *Embedding) MarshalBinary() ([]byte, error) {
	return emb.Serialize(), nil
}

func (emb *Embedding) Serialize() []byte {
	size := 8 + len(emb.SG.V)*8 + len(emb.SG.E)*12
	label := make([]byte, size)
	binary.BigEndian.PutUint32(label[0:4], uint32(len(emb.SG.E)))
	binary.BigEndian.PutUint32(label[4:8], uint32(len(emb.SG.V)))
	off := 8
	for i := range emb.SG.V {
		s := off + i*8
		e := s + 4
		binary.BigEndian.PutUint32(label[s:e], uint32(emb.Ids[i]))
		s += 4
		e += 4
		binary.BigEndian.PutUint32(label[s:e], uint32(emb.SG.V[i].Color))
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
	return emb.SG.Label()
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

func (emb *Embedding) Pretty(colors []string) string {
	V := make([]string, 0, len(emb.SG.V))
	E := make([]string, 0, len(emb.SG.E))
	for i := range emb.SG.V {
		V = append(V, fmt.Sprintf(
			"(%v:%v)",
			emb.Ids[i],
			colors[emb.SG.V[i].Color],
		))
	}
	for i := range emb.SG.E {
		E = append(E, fmt.Sprintf(
			"[%v->%v:%v]",
			emb.SG.E[i].Src,
			emb.SG.E[i].Targ,
			colors[emb.SG.E[i].Color],
		))
	}
	return fmt.Sprintf("{%v:%v}%v%v", len(emb.SG.E), len(emb.SG.V), strings.Join(V, ""), strings.Join(E, ""))
}

func (emb *Embedding) Dotty(G *goiso.Graph, attrs map[int]map[string]interface{}) string {
	V := make([]string, 0, len(emb.SG.V))
	E := make([]string, 0, len(emb.SG.E))
	safeStr := func(i interface{}) string {
		s := fmt.Sprint(i)
		s = strings.Replace(s, "\n", "\\n", -1)
		s = strings.Replace(s, "\"", "\\\"", -1)
		return s
	}
	renderAttrs := func(color, id int) string {
		a := attrs[id]
		label := G.Colors[color]
		strs := make([]string, 0, len(a)+1)
		strs = append(strs, fmt.Sprintf(`idx="%v"`, id))
		if line, has := a["start_line"]; has {
			strs = append(strs, fmt.Sprintf(`label="%v\n[line: %v]"`, safeStr(label), safeStr(line)))
		} else {
			strs = append(strs, fmt.Sprintf(`label="%v"`, safeStr(label)))
		}
		for name, value := range a {
			if name == "label" || name == "id" {
				continue
			}
			strs = append(strs, fmt.Sprintf("%v=\"%v\"", name, safeStr(value)))
		}
		return strings.Join(strs, ",")
	}
	for idx, id := range emb.Ids {
		V = append(V, fmt.Sprintf(
			"%v [%v];",
			G.V[id].Id,
			renderAttrs(emb.SG.V[idx].Color, id),
		))
	}
	for idx := range emb.SG.E {
		e := &emb.SG.E[idx]
		E = append(E, fmt.Sprintf(
			"%v -> %v [label=\"%v\"];",
			G.V[emb.Ids[e.Src]].Id,
			G.V[emb.Ids[e.Targ]].Id,
			safeStr(G.Colors[e.Color]),
		))
	}
	return fmt.Sprintf(
		`digraph {
    %v
    %v
}
`, strings.Join(V, "\n    "), strings.Join(E, "\n    "))
}
