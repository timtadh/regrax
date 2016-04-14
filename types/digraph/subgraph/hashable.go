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
