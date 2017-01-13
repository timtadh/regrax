package subgraph

import (
	"github.com/timtadh/data-structures/types"
)

type Extension struct {
	Source Vertex
	Target Vertex
	Color  int
}

func NewExt(src, targ Vertex, color int) *Extension {
	return &Extension{
		Source: src,
		Target: targ,
		Color:  color,
	}
}

func (e *Extension) Translate(orgLen int, vord []int) *Extension {
	srcIdx := e.Source.Idx
	targIdx := e.Target.Idx
	if srcIdx >= orgLen {
		srcIdx = len(vord) + (srcIdx - orgLen)
	}
	if targIdx >= orgLen {
		targIdx = len(vord) + (targIdx - orgLen)
	}
	if srcIdx < len(vord) {
		srcIdx = vord[srcIdx]
	}
	if targIdx < len(vord) {
		targIdx = vord[targIdx]
	}
	return &Extension{
		Source: Vertex{
			Idx:   srcIdx,
			Color: e.Source.Color,
		},
		Target: Vertex{
			Idx:   targIdx,
			Color: e.Target.Color,
		},
		Color: e.Color,
	}
}

func (e *Extension) Equals(o types.Equatable) bool {
	switch x := o.(type) {
	case *Extension:
		return e.Source.Idx == x.Source.Idx &&
			e.Source.Color == x.Source.Color &&
			e.Target.Idx == x.Target.Idx &&
			e.Target.Color == x.Target.Color &&
			e.Color == x.Color
	}
	return false
}

func (e *Extension) Less(o types.Sortable) bool {
	switch x := o.(type) {
	case *Extension:
		if e.Source.Idx < x.Source.Idx {
			return true
		} else if e.Source.Idx > x.Source.Idx {
			return false
		}
		if e.Source.Color < x.Source.Color {
			return true
		} else if e.Source.Color > x.Source.Color {
			return false
		}
		if e.Target.Idx < x.Target.Idx {
			return true
		} else if e.Target.Idx > x.Target.Idx {
			return false
		}
		if e.Target.Color < x.Target.Color {
			return true
		} else if e.Target.Color > x.Target.Color {
			return false
		}
		if e.Color < x.Color {
			return true
		}
		return false
	}
	return false
}

func (e *Extension) Hash() int {
	return e.Source.Idx +
		2*e.Source.Color +
		3*e.Target.Idx +
		5*e.Target.Color +
		7*e.Color
}
