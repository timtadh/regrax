package graph

import (
	"bytes"
	"fmt"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/fs2"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_int"
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

func (sgs SubGraphs) Verify() {
	label := sgs[0].ShortLabel()
	for _, sg := range sgs {
		if !bytes.Equal(label, sg.ShortLabel()) {
			panic(fmt.Errorf("bad partition %v %v", sgs[0].Label(), sg.Label()))
		}
	}
}

func (sgs SubGraphs) Partition() []SubGraphs {
	sort.Sort(sgs)
	parts := make([]SubGraphs, 0, 10)
	buf := make(SubGraphs, 0, 10)
	var ckey []byte = nil
	for _, sg := range sgs {
		label := sg.ShortLabel()
		if ckey != nil && !bytes.Equal(ckey, label) {
			parts = append(parts, buf)
			buf = make(SubGraphs, 0, 10)
		}
		ckey = label
		buf = append(buf, sg)
	}
	if len(buf) > 0 {
		parts = append(parts, buf)
	}
	for _, part := range parts {
		part.Verify()
	}
	return parts
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
	if len(n.sgs) > 0 {
		return fmt.Sprintf("<Node %v %v>", len(n.sgs), n.sgs[0].Label())
	} else {
		return fmt.Sprintf("<Node %v {}>", len(n.sgs))
	}
}

func (n *Node) StartingPoint() bool {
	return n.Size() == 1
}

func (n *Node) Size() int {
	return len(n.sgs[0].E)
}

func (n *Node) Parents(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	dt := dtype.(*Graph)
	if len(n.sgs) == 0 {
		return []lattice.Node{}, nil
	}
	if len(n.sgs[0].V) == 1 && len(n.sgs[0].E) == 0 {
		return []lattice.Node{&Node{}}, nil
	}
	if nodes, has, err := n.cached(dt, dt.ParentCount, dt.Parents, n.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	parents := make([]lattice.Node, 0, 10)
	for _, parent := range n.sgs[0].SubGraphs() {
		p, err := n.Parent(support, dt, parent)
		if err != nil {
			return nil, err
		}
		parents = append(parents, p)
	}
	return parents, n.cache(dt, dt.ParentCount, dt.Parents, n.label, parents)
}

func (n *Node) Parent(support int, dt *Graph, target *goiso.SubGraph) (*Node, error) {
	// errors.Logf("DEBUG", "target %v", target.Label())
	csg, cur, err := n.parentStart(dt, target)
	if err != nil {
		return nil, err
	}
	cur.Verify()
	// errors.Logf("DEBUG", "start %v", cur[0].Label())
	dequeue := func(queue []*goiso.Edge) ([]*goiso.Edge, *goiso.Edge) {
		e := queue[0]
		copy(queue[0:len(queue)-1], queue[1:])
		return queue[0:len(queue)-1], e
	}
	queue := make([]*goiso.Edge, 0, len(target.E))
	for i := range target.E {
		queue = append(queue, &target.E[i])
	}
	unchanged := 0
	for len(queue) > 0 {
		var e *goiso.Edge
		queue, e = dequeue(queue)
		if unchanged > len(queue) + 1 {
			queue = append(queue, e)
			errors.Logf("ERROR", "cannot find any of the edges %v for target %v\n cur %v", len(queue), target, cur[0])
			return nil, errors.Errorf("cannot find any of the edges %v for target %v", len(queue), target)
		} else if nsg := cur.nextSg(target, csg, e); nsg != nil {
			errors.Logf("DEBUG", "extend to %v", nsg.Label())
			next, err := cur.extendWith(dt, target, nsg, e)
			if err != nil {
				return nil, err
			}
			if len(next) <= 0 {
				queue = append(queue, e)
				unchanged++
			} else {
				csg = nsg
				cur = next
				cur.Verify()
				unchanged = 0
			}
		} else {
			queue = append(queue, e)
			unchanged++
		}
	}
	if len(cur) - 1 < support {
		return &Node{cur[0].ShortLabel(), MaxIndepSupported(cur)}, nil
	} else {
		return &Node{cur[0].ShortLabel(), MinImgSupported(cur)}, nil
	}
}

func (n *Node) parentStart(dt *Graph, target *goiso.SubGraph) (*goiso.SubGraph, SubGraphs, error) {
	start := dt.G.SubGraph([]int{target.V[0].Id}, nil)
	startLabel := start.ShortLabel()
	var sgs SubGraphs
	err := dt.Embeddings.DoFind(startLabel, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return start, sgs, nil
}

func (sgs SubGraphs) nextSg(target, csg *goiso.SubGraph, e *goiso.Edge) (nsg *goiso.SubGraph) {
	src := target.V[e.Src]
	targ := target.V[e.Targ]
	if !(csg.HasVertex(src.Id) || csg.HasVertex(targ.Id)) {
		return nil
	}
	nsg = csg.EdgeExtend(&goiso.Edge{
		Color: e.Color,
		Arc: goiso.Arc{
			Src: src.Id,
			Targ: targ.Id,
		},
	})
	return nsg
}

func (sgs SubGraphs) extendWith(dt *Graph, target, ref *goiso.SubGraph, e *goiso.Edge) (SubGraphs, error) {
	src := target.V[e.Src]
	targ := target.V[e.Targ]
	mkexts := func(sg *goiso.SubGraph) (exts []*goiso.SubGraph, err error) {
		for _, u := range sg.V {
			if u.Color == src.Color {
				for _, x := range dt.G.Kids[u.Id] {
					if dt.G.V[x.Targ].Color != targ.Color {
						continue
					} else if x.Color != e.Color {
						continue
					}
					exts = append(exts, sg.EdgeExtend(x))
				}
			} else if u.Color == targ.Color {
				for _, x := range dt.G.Parents[u.Id] {
					if dt.G.V[x.Src].Color != src.Color {
						continue
					} else if x.Color != e.Color {
						continue
					}
					exts = append(exts, sg.EdgeExtend(x))
				}
			}
		}
		return exts, nil
	}
	refLabel := ref.ShortLabel()
	exts := make(SubGraphs, 0, len(sgs))
	for _, sg := range sgs {
		es, err := mkexts(sg)
		if err != nil {
			return nil, err
		}
		for _, ext := range es {
			errors.Logf("DEBUG", "\n    ref %v\n    new %v", ref.Label(), ext.Label())
			if !bytes.Equal(refLabel, ext.ShortLabel()) {
				continue
			}
			exts = append(exts, ext)
			// errors.Logf("DEBUG", "added %v", ext.Label())
		}
	}
	return DedupSupported(exts), nil
}

func (n *Node) Children(support int, dtype lattice.DataType) (nodes []lattice.Node, err error) {
	dt := dtype.(*Graph)
	if len(n.sgs) == 0 {
		return dt.FrequentVertices, nil
	}
	if nodes, has, err := n.cached(dt, dt.ChildCount, dt.Children, n.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
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
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, n.cache(dt, dt.ChildCount, dt.Children, n.label, nodes)
}

func (n *Node) cache(dt *Graph, count bytes_int.MultiMap, cache fs2.MultiMap, key []byte, nodes []lattice.Node) (err error) {
	if has, err := count.Has(key); err != nil {
		return err
	} else if has {
		return nil
	}
	err = count.Add(key, int32(len(nodes)))
	if err != nil {
		return err
	}
	for _, node := range nodes {
		err = node.(*Node).Save(dt)
		if err != nil {
			return err
		}
		err = cache.Add(key, node.(*Node).label)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) cached(dt *Graph, count bytes_int.MultiMap, cache fs2.MultiMap, key []byte) (nodes []lattice.Node, has bool, err error) {
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
	err = cache.DoFind(key, func(_, adj []byte) error {
		sgs := make(SubGraphs, 0, 10)
		err := dt.Embeddings.DoFind(adj, func(_ []byte, sg *goiso.SubGraph) error {
			sgs = append(sgs, sg)
			return nil
		})
		if err != nil {
			return err
		}
		nodes = append(nodes, &Node{adj, sgs})
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return nodes, true, nil
}

func (n *Node) AdjacentCount(support int, dtype lattice.DataType) (int, error) {
	pc, err := n.ParentCount(support, dtype)
	if err != nil {
		return 0, err
	}
	cc, err := n.ChildCount(support, dtype)
	if err != nil {
		return 0, err
	}
	return pc + cc, nil
}

func (n *Node) ParentCount(support int, dtype lattice.DataType) (int, error) {
	dt := dtype.(*Graph)
	if len(n.sgs) == 0 {
		return 0, nil
	}
	if has, err := dt.ParentCount.Has(n.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Parents(support, dt)
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := dt.ParentCount.DoFind(n.label, func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) ChildCount(support int, dtype lattice.DataType) (int, error) {
	dt := dtype.(*Graph)
	if len(n.sgs) == 0 {
		return len(dt.FrequentVertices), nil
	}
	if has, err := dt.ChildCount.Has(n.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Children(support, dt)
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := dt.ChildCount.DoFind(n.label, func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) Maximal(support int, dtype lattice.DataType) (bool, error) {
	cc, err := n.ChildCount(support, dtype)
	if err != nil {
		return false, err
	}
	return cc == 0, nil
}

func (n *Node) Label() []byte {
	return n.label
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

