package graph

import (
	"bytes"
	"fmt"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
)



type Node struct {
	label []byte
	sgs []*goiso.SubGraph
}

type Embedding struct {
	sg *goiso.SubGraph
}

type SubGraphs []*goiso.SubGraph

func (sgs SubGraphs) Len() int {
	return len(sgs)
}

func (sgs SubGraphs) Less(i, j int) bool {
	return bytes.Compare(sgs[i].ShortLabel(), sgs[j].ShortLabel()) < 0
}

func (sgs SubGraphs) Swap(i, j int) {
	sgs[i], sgs[j] = sgs[j], sgs[i]
}

func (sgs SubGraphs) Partition() []SubGraphs {
	sort.Sort(sgs)
	part := make([]SubGraphs, 0, 10)
	buf := make(SubGraphs, 0, 10)
	var ckey []byte = nil
	for _, sg := range sgs {
		label := sg.ShortLabel()
		if ckey != nil && !bytes.Equal(ckey, label) {
			part = append(part, buf)
			buf = make(SubGraphs, 0, 10)
		}
		ckey = label
		buf = append(buf, sg)
	}
	if len(buf) > 0 {
		part = append(part, buf)
	}
	return part
}


func (n *Node) Save(dt *Graph) error {
	return errors.Errorf("unimplemented")
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.sgs[0].Label(), len(n.sgs))
}

func (n *Node) StartingPoint() bool {
	return n.Size() == 1
}

func (n *Node) Size() int {
	return len(n.sgs[0].E)
}

func (n *Node) Parents(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *Node) Children(support int, dtype lattice.DataType) (nodes []lattice.Node, err error) {
	dt := dtype.(*Graph)
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) SubGraphs {
		if dt.G.ColorFrequency(e.Color) < support {
			return exts
		} else if dt.G.ColorFrequency(dt.G.V[e.Src].Color) < support {
			return exts
		} else if dt.G.ColorFrequency(dt.G.V[e.Targ].Color) < support {
			return exts
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			exts = append(exts, sg.EdgeExtend(e))
		}
		return exts
	}
	for _, sg := range n.sgs {
		for _, u := range sg.V {
			for _, e := range dt.G.Kids[u.Id] {
				exts = add(exts, sg, e)
			}
			for _, e := range dt.G.Parents[u.Id] {
				exts = add(exts, sg, e)
			}
		}
	}
	partitioned := exts.Partition()
	for _, sgs := range partitioned {
		sgs = MinImgSupported(sgs)
		if len(sgs) >= support {
			label := sgs[0].ShortLabel()
			nodes = append(nodes, &Node{label, sgs})
		}
	}
	return nodes, nil
}

func (n *Node) AdjacentCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) ParentCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) ChildCount(support int, dtype lattice.DataType) (int, error) {
	return 0, errors.Errorf("unimplemented")
}

func (n *Node) Maximal(support int, dtype lattice.DataType) (bool, error) {
	return false, errors.Errorf("unimplemented")
}

func (n *Node) Label() []byte {
	return nil
}

func (n *Node) Embeddings() ([]lattice.Embedding, error) {
	return nil, errors.Errorf("unimplemented")
}

func (n *Node) Lattice(support int, dtype lattice.DataType) (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (e *Embedding) Components() ([]int, error) {
	return nil, errors.Errorf("unimplemented")
}

