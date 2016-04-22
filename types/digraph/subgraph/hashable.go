package subgraph

import (
	"github.com/timtadh/data-structures/types"
)

func (sg *SubGraph) Equals(o types.Equatable) bool {
	a := types.ByteSlice(sg.Label())
	switch b := o.(type) {
	case *SubGraph:
		return a.Equals(types.ByteSlice(b.Label()))
	default:
		return false
	}
}

func (sg *SubGraph) Less(o types.Sortable) bool {
	a := types.ByteSlice(sg.Label())
	switch b := o.(type) {
	case *SubGraph:
		return a.Less(types.ByteSlice(b.Label()))
	default:
		return false
	}
}

func (sg *SubGraph) Hash() int {
	return types.ByteSlice(sg.Label()).Hash()
}

func (emb *Embedding) Equals(o types.Equatable) bool {
	a := types.ByteSlice(emb.Serialize())
	switch b := o.(type) {
	case *Embedding:
		return a.Equals(types.ByteSlice(b.Serialize()))
	default:
		return false
	}
}

func (emb *Embedding) Less(o types.Sortable) bool {
	a := types.ByteSlice(emb.Serialize())
	switch b := o.(type) {
	case *Embedding:
		return a.Less(types.ByteSlice(b.Serialize()))
	default:
		return false
	}
}

func (emb *Embedding) Hash() int {
	return types.ByteSlice(emb.Serialize()).Hash()
}

