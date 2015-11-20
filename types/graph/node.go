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
	if has, err := dt.Embeddings.Has(n.label); err != nil {
		return err
	} else if has {
		return nil
	}
	for _, sg := range n.sgs {
		err := dt.Embeddings.Add(n.label, sg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", len(n.sgs), n.sgs[0].Label())
}

func (n *Node) StartingPoint() bool {
	return n.Size() == 1
}

func (n *Node) Size() int {
	return len(n.sgs[0].E)
}

func (n *Node) Parents(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	dt := dtype.(*Graph)
	parents := make([]lattice.Node, 0, 10)
	for _, parent := range n.sgs[0].SubGraphs() {
		p, err := n.Parent(support, dt, parent)
		if err != nil {
			return nil, err
		}
		parents = append(parents, p)
	}
	return parents, nil
}

func (n *Node) Parent(support int, dt *Graph, target *goiso.SubGraph) (*Node, error) {
	errors.Logf("DEBUG", "target %v", target.Label())
	cur, err := n.parentStart(dt, target)
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "start %v", cur)
	dequeue := func(queue []*goiso.Edge) ([]*goiso.Edge, *goiso.Edge) {
		e := queue[0]
		copy(queue[0:len(queue)-1], queue[1:])
		return queue[0:len(queue)-1], e
	}
	queue := make([]*goiso.Edge, 0, len(target.E))
	for i := range target.E {
		queue = append(queue, &target.E[i])
	}
	for len(queue) > 0 {
		var e *goiso.Edge
		queue, e = dequeue(queue)
		if s, t, found := cur.findEdge(target, e); found {
			cur, err = cur.extendWith(dt, target, e, s, t)
			if err != nil {
				return nil, err
			}
			errors.Logf("DEBUG", "cur %v", cur)
		} else {
			queue = append(queue, e)
		}
	}
	return cur, nil
}

func (n *Node) parentStart(dt *Graph, target *goiso.SubGraph) (*Node, error) {
	startLabel := dt.G.SubGraph([]int{target.V[target.E[0].Src].Id}, nil).ShortLabel()
	var sgs SubGraphs
	err := dt.Embeddings.DoFind(startLabel, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Node{startLabel, sgs}, nil
}

func (n *Node) findEdge(target *goiso.SubGraph, e *goiso.Edge) (srcIdx, targIdx int, found bool) {
	src := target.V[e.Src]
	targ := target.V[e.Targ]
	for _, sg := range n.sgs {
		for idx, u := range sg.V {
			if src.Id == u.Id {
				return idx, -1, true
			} else if targ.Id == u.Id {
				return -1, idx, true
			}
		}
	}
	return -1, -1, false
}

func (n *Node) extendWith(dt *Graph, target *goiso.SubGraph, e *goiso.Edge, srcIdx, targIdx int) (*Node, error) {
	src := target.V[e.Src]
	targ := target.V[e.Targ]
	sgs := make(SubGraphs, 0, len(n.sgs))
	for _, sg := range n.sgs {
		if srcIdx >= 0 {
			u := sg.V[srcIdx]
			for _, x := range dt.G.Kids[u.Id] {
				if dt.G.V[x.Targ].Color != targ.Color {
					continue
				} else if x.Color != e.Color {
					continue
				}
				sgs = append(sgs, sg.EdgeExtend(x))
			}
		} else if targIdx >= 0{
			u := sg.V[targIdx]
			for _, x := range dt.G.Parents[u.Id] {
				if dt.G.V[x.Src].Color != src.Color {
					continue
				} else if x.Color != e.Color {
					continue
				}
				sgs = append(sgs, sg.EdgeExtend(x))
			}
		} else {
			return nil, errors.Errorf("Could not find the edge %v from %v in %v", e, target.Label(), n)
		}
	}
	sgs = MinImgSupported(sgs)
	return &Node{sgs[0].ShortLabel(), sgs}, nil
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
	for _, node := range nodes {
		err := node.(*Node).Save(dt)
		if err != nil {
			return nil, err
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

