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
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)



type Node struct {
	dt *Graph
	label []byte
	sgs SubGraphs
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


func (n *Node) Save() error {
	if has, err := n.dt.Embeddings.Has(n.label); err != nil {
		return err
	} else if has {
		return nil
	}
	for _, sg := range n.sgs {
		err := n.dt.Embeddings.Add(n.label, sg)
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
	if len(n.sgs) > 0 {
		return len(n.sgs[0].E)
	}
	return 0
}

func (n *Node) Parents() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "compute Parents\n    of %v", n)
	if len(n.sgs) == 0 {
		return []lattice.Node{}, nil
	}
	if len(n.sgs[0].V) == 1 && len(n.sgs[0].E) == 0 {
		return []lattice.Node{&Node{dt: n.dt}}, nil
	}
	if nodes, has, err := n.cached(n.dt.ParentCount, n.dt.Parents, n.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	parents := make([]lattice.Node, 0, 10)
	for _, parent := range n.sgs[0].SubGraphs() {
		p, err := n.FindNode(n.dt, parent)
		if err != nil {
			return nil, err
		}
		if p != nil {
			parents = append(parents, p)
		}
	}
	if len(parents) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return parents, n.cache(n.dt.ParentCount, n.dt.Parents, n.label, parents)
}

func (n *Node) FindNode(dt *Graph, target *goiso.SubGraph) (*Node, error) {
	label := target.ShortLabel()
	if has, err := n.dt.Embeddings.Has(label); err != nil {
		return nil, err
	} else if has {
		sgs := make(SubGraphs, 0, 10)
		err := n.dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
			sgs = append(sgs, sg)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &Node{n.dt, label, sgs}, nil
	}
	// errors.Logf("DEBUG", "target %v", target.Label())
	// errors.Logf("DEBUG", "compute Parent\n    of %v", target.Label())
	cur, graphs, err := edgeChain(n.dt, target)
	if err != nil {
		return nil, err
	}
	cur.sgs.Verify()
	for _, sg := range graphs {
		// errors.Logf("DEBUG", "\n    extend %v\n        to %v", cur.sgs[0].Label(), sg.Label())
		cur, err = cur.extendTo(sg)
		if err != nil {
			return nil, err
		}
		if cur == nil {
			return nil, nil
		}
		cur.sgs.Verify()
	}
	return cur, cur.Save()
}

func edgeChain(dt *Graph, target *goiso.SubGraph) (start *Node, graphs []*goiso.SubGraph, err error) {
	cur := target
	// errors.Logf("DEBUG", "doing edgeChain\n    of %v", target.Label())
	graphs = make([]*goiso.SubGraph, len(cur.E))
	for chainIdx := len(graphs) - 1; chainIdx >= 0; chainIdx-- {
		if len(cur.V) <= 2 && len(cur.E) <= 1 {
			a := cur.G.VertexSubGraph(cur.V[0].Id)
			graphs[chainIdx] = cur
			// errors.Logf("DEBUG", "small parent a %v cur %v", a.Label(), cur.Label())
			cur = a
		} else {
			for i := range cur.E {
				p := cur.RemoveEdge(i)
				if p.Connected() {
					graphs[chainIdx] = cur
					cur = p
					break
				}
			}
		}
		// errors.Logf("DEBUG", "rgraph %v %v", chainIdx, graphs[chainIdx].Label())
		if graphs[chainIdx] == nil {
			return nil, nil, errors.Errorf("Could not find a connected parent!")
		}
	}
	// for i, g := range graphs {
		// errors.Logf("DEBUG", "graph %v %v", i, g.Label())
	// }
	startSg := dt.G.VertexSubGraph(cur.V[0].Id)
	startLabel := startSg.ShortLabel()
	var sgs SubGraphs
	err = dt.Embeddings.DoFind(startLabel, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &Node{dt, startLabel, sgs}, graphs, nil
}

func (n *Node) extendTo(sg *goiso.SubGraph) (exts *Node, err error) {
	// errors.Logf("DEBUG", "doing extendTo\n    extend %v\n        to %v", n.sgs[0].Label(), sg.Label())
	latKids, err := n.Children()
	if err != nil {
		return nil, err
	}
	label := sg.ShortLabel()
	for _, lkid := range latKids {
		kid := lkid.(*Node)
		// errors.Logf("DEBUG", "trying to find extension\n    from %v\n      to %v\n    have %v", n.sgs[0].Label(), sg.Label(), kid.sgs[0].Label())
		if bytes.Equal(label, kid.label) {
			// errors.Logf("DEBUG", "FOUND extension\n    from %v\n      to %v\n    have %v", n.sgs[0].Label(), sg.Label(), kid.sgs[0].Label())
			return kid, nil
		}
	}
	// errors.Logf("DEBUG", "could not find extension\n    from %v\n      to %v", n.sgs[0].Label(), sg.Label())
	return nil, nil
}

func (n *Node) Children() (nodes []lattice.Node, err error) {
	if len(n.sgs) == 0 {
		return n.dt.FrequentVertices, nil
	}
	if len(n.sgs[0].E) >= n.dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := n.cached(n.dt.ChildCount, n.dt.Children, n.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) SubGraphs {
		if n.dt.G.ColorFrequency(e.Color) < n.dt.Support() {
			return exts
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Src].Color) < n.dt.Support() {
			return exts
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Targ].Color) < n.dt.Support() {
			return exts
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			ext := sg.EdgeExtend(e)
			if len(ext.V) <= n.dt.MaxVertices {
				exts = append(exts, ext)
			}
		}
		return exts
	}
	for _, sg := range n.sgs {
		for _, u := range sg.V {
			for _, e := range n.dt.G.Kids[u.Id] {
				exts = add(exts, sg, e)
			}
			for _, e := range n.dt.G.Parents[u.Id] {
				exts = add(exts, sg, e)
			}
		}
	}
	partitioned := exts.Partition()
	for _, sgs := range partitioned {
		sgs = MinImgSupported(sgs)
		if len(sgs) >= n.dt.Support() {
			label := sgs[0].ShortLabel()
			nodes = append(nodes, &Node{n.dt, label, sgs})
		}
	}
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, n.cache(n.dt.ChildCount, n.dt.Children, n.label, nodes)
}

func (n *Node) cache(count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte, nodes []lattice.Node) (err error) {
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
		err = node.(*Node).Save()
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

func (n *Node) cached(count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte) (nodes []lattice.Node, has bool, err error) {
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
	err = cache.DoFind(key, func(_, adj []byte) error {
		sgs := make(SubGraphs, 0, 10)
		err := n.dt.Embeddings.DoFind(adj, func(_ []byte, sg *goiso.SubGraph) error {
			sgs = append(sgs, sg)
			return nil
		})
		if err != nil {
			return err
		}
		nodes = append(nodes, &Node{n.dt, adj, sgs})
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return nodes, true, nil
}

func (n *Node) AdjacentCount() (int, error) {
	pc, err := n.ParentCount()
	if err != nil {
		return 0, err
	}
	cc, err := n.ChildCount()
	if err != nil {
		return 0, err
	}
	return pc + cc, nil
}

func (n *Node) ParentCount() (int, error) {
	if len(n.sgs) == 0 {
		return 0, nil
	}
	if has, err := n.dt.ParentCount.Has(n.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Parents()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ParentCount.DoFind(n.label, func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) ChildCount() (int, error) {
	if len(n.sgs) == 0 {
		return len(n.dt.FrequentVertices), nil
	}
	if has, err := n.dt.ChildCount.Has(n.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Children()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ChildCount.DoFind(n.label, func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) Maximal() (bool, error) {
	cc, err := n.ChildCount()
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

func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (e *Embedding) Components() ([]int, error) {
	return nil, errors.Errorf("unimplemented")
}

