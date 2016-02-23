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
	parents, err := allParents(ext)
	if err != nil {
		return false, err
	} else if len(parents) == 0 {
		return false, errors.Errorf("ext %v of node %v has no parents", ext.Label(), cur.Label())
	}
	parent := parents[0]
	if bytes.Equal(parent.ShortLabel(), cur.ShortLabel()) {
		return true, nil
	}
	return false, nil
	// this check is really about seeing if the canon parent is reliably reachable
	// however that should always be the case. In some instances it is not because
	// of the choices made in which subgraphs are kep
	// p, err := FindNode(n.dt, parent)
	// if err != nil {
	// 	return false, err
	// } else if p != nil {
	// 	return false, nil
	// }
	// return false, errors.Errorf("could not find parent %v of extention %v for node %v", parent.Label(), ext.Label(), n)
}

func allParents(sg *goiso.SubGraph) ([]*goiso.SubGraph, error) {
	if len(sg.E) <= 0 {
		return nil, nil
	}
	parents := make([]*goiso.SubGraph, 0, 10)
	for i := len(sg.E)-1; i >= 0; i-- {
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
	}
	return parents, nil
}

