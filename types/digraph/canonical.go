package digraph

import (
	"bytes"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
)

func isCanonicalExtension(cur *goiso.SubGraph, ext *goiso.SubGraph) (bool, error) {
	// errors.Logf("DEBUG", "is %v a canonical ext of %v", ext.Label(), n)
	parent, err := firstParent(ext)
	if err != nil {
		return false, err
	} else if parent == nil {
		return false, errors.Errorf("ext %v of node %v has no parents", ext.Label(), cur.Label())
	}
	if bytes.Equal(parent.ShortLabel(), cur.ShortLabel()) {
		return true, nil
	}
	return false, nil
}

func computeParent(sg *goiso.SubGraph, i int, parents []*goiso.SubGraph) ([]*goiso.SubGraph) {
	if len(sg.V) == 2 && len(sg.E) == 1 {
		p, _ := sg.G.VertexSubGraph(sg.V[sg.E[0].Src].Id)
		parents = append(parents, p)
	} else if len(sg.V) == 1 && len(sg.E) == 1 {
		p, _ := sg.G.VertexSubGraph(sg.V[sg.E[0].Src].Id)
		parents = append(parents, p)
		p, _ = sg.G.VertexSubGraph(sg.V[sg.E[0].Targ].Id)
		parents = append(parents, p)
	} else {
		p, _ := sg.RemoveEdge(i)
		if p.Connected() {
			parents = append(parents, p)
		}
	}
	return parents
}

func firstParent(sg *goiso.SubGraph) (*goiso.SubGraph, error) {
	if len(sg.E) <= 0 {
		return nil, nil
	}
	parents := make([]*goiso.SubGraph, 0, 10)
	for i := len(sg.E)-1; i >= 0; i-- {
		parents = computeParent(sg, i, parents)
		if len(parents) > 0 {
			return parents[0], nil
		}
	}
	return nil, errors.Errorf("no parents for %v", sg.Label())
}


func allParents(sg *goiso.SubGraph) ([]*goiso.SubGraph, error) {
	if len(sg.E) <= 0 {
		return nil, nil
	}
	parents := make([]*goiso.SubGraph, 0, 10)
	for i := len(sg.E)-1; i >= 0; i-- {
		computeParent(sg, i, parents)
	}
	return parents, nil
}

