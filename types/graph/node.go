package graph

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

type GraphPattern struct {
	label []byte
	level int
	sg *goiso.SubGraph
}

var EmptyPattern = &GraphPattern{
	label: []byte{},
	level: 0,
	sg: nil,
}

type Node struct {
	GraphPattern
	dt    *Graph
	sgs   SubGraphs
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

func (sgs SubGraphs) Verify() error {
	if len(sgs) <= 0 {
		return errors.Errorf("empty partition")
	}
	label := sgs[0].ShortLabel()
	for _, sg := range sgs {
		if !bytes.Equal(label, sg.ShortLabel()) {
			return errors.Errorf("bad partition %v %v", sgs[0].Label(), sg.Label())
		}
	}
	return nil
}

func (sgs SubGraphs) Partition() []SubGraphs {
	sort.Sort(sgs)
	parts := make([]SubGraphs, 0, 10)
	add := func(parts []SubGraphs, buf SubGraphs) []SubGraphs {
		err := buf.Verify()
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		} else {
			parts = append(parts, buf)
		}
		return parts
	}
	buf := make(SubGraphs, 0, 10)
	var ckey []byte = nil
	for _, sg := range sgs {
		label := sg.ShortLabel()
		if ckey != nil && !bytes.Equal(ckey, label) {
			parts = add(parts, buf)
			buf = make(SubGraphs, 0, 10)
		}
		ckey = label
		buf = append(buf, sg)
	}
	if len(buf) > 0 {
		parts = add(parts, buf)
	}
	return parts
}

func (n *Node) Pattern() lattice.Pattern {
	n.GraphPattern.level = n.Level()
	if n.GraphPattern.level > 0 {
		n.GraphPattern.sg = n.sgs[0]
	}
	return &n.GraphPattern
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

func (n *Node) Level() int {
	if len(n.sgs) > 0 {
		return len(n.sgs[0].E) + 1
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
			errors.Logf("ERROR", "%v", err)
		} else if p != nil {
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
		return &Node{GraphPattern{label: label}, n.dt, sgs}, nil
	}
	// errors.Logf("DEBUG", "target %v", target.Label())
	// errors.Logf("DEBUG", "compute Parent\n    of %v", target.Label())
	cur, graphs, err := edgeChain(n.dt, target)
	if err != nil {
		return nil, err
	}
	err = cur.sgs.Verify()
	if err != nil {
		return nil, err
	}
	for _, sg := range graphs {
		// errors.Logf("DEBUG", "\n    extend %v\n        to %v", cur.sgs[0].Label(), sg.Label())
		cur, err = cur.extendTo(sg)
		if err != nil {
			return nil, err
		}
		if cur == nil {
			return nil, nil
		}
		err := cur.sgs.Verify()
		if err != nil {
			return nil, err
		}
	}
	return cur, cur.Save()
}

func edgeChain(dt *Graph, target *goiso.SubGraph) (start *Node, graphs []*goiso.SubGraph, err error) {
	cur := target
	// errors.Logf("DEBUG", "doing edgeChain\n    of %v", target.Label())
	graphs = make([]*goiso.SubGraph, len(cur.E))
	for chainIdx := len(graphs) - 1; chainIdx >= 0; chainIdx-- {
		if len(cur.V) <= 2 && len(cur.E) <= 1 {
			a, _ := cur.G.VertexSubGraph(cur.V[0].Id)
			graphs[chainIdx] = cur
			// errors.Logf("DEBUG", "small parent a %v cur %v", a.Label(), cur.Label())
			cur = a
		} else {
			for i := range cur.E {
				p, _ := cur.RemoveEdge(i)
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
	startSg, _ := dt.G.VertexSubGraph(cur.V[0].Id)
	startLabel := startSg.ShortLabel()
	var sgs SubGraphs
	err = dt.Embeddings.DoFind(startLabel, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &Node{GraphPattern{label: startLabel}, dt, sgs}, graphs, nil
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
	return n.children(false)
}

func (n *Node) CanonKids() (nodes []lattice.Node, err error) {
	return n.children(true)
}

func (n *Node) children(checkCanon bool) (nodes []lattice.Node, err error) {
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
			ext, canonized := sg.EdgeExtend(e)
			if len(ext.V) <= n.dt.MaxVertices && (!checkCanon || canonized) {
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
		sgs = MinImgSupported(DedupSupported(sgs))
		if len(sgs) >= n.dt.Support() {
			label := sgs[0].ShortLabel()
			nodes = append(nodes, &Node{GraphPattern{label: label}, n.dt, sgs})
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
		nodes = append(nodes, &Node{GraphPattern{label: adj}, n.dt, sgs})
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

func NewGraphPattern(sg *goiso.SubGraph) *GraphPattern {
	return &GraphPattern{
		label: sg.ShortLabel(),
		level: len(sg.E) + 1,
		sg: sg,
	}
}

func (g *GraphPattern) Label() []byte {
	return g.label
}

func (g *GraphPattern) Level() int {
	return g.level
}

func (g *GraphPattern) String() string {
	if g.sg != nil {
		return fmt.Sprintf("<GraphPattern %v>", g.sg.Label())
	} else {
		return fmt.Sprintf("<GraphPattern {}>")
	}
}

func (a *GraphPattern) Distance(o lattice.Pattern) float64 {
	b := o.(*GraphPattern)
	aColors := set.NewSortedSet(len(a.sg.V) + len(a.sg.E))
	bColors := set.NewSortedSet(len(b.sg.V) + len(b.sg.E))
	for i := range a.sg.V {
		aColors.Add(types.Int(a.sg.V[i].Color))
	}
	for i := range b.sg.V {
		bColors.Add(types.Int(b.sg.V[i].Color))
	}
	for i := range a.sg.E {
		aColors.Add(types.Int(a.sg.E[i].Color))
	}
	for i := range b.sg.E {
		bColors.Add(types.Int(b.sg.E[i].Color))
	}
	intersect, _ := aColors.Intersect(bColors)
	overlap := float64(intersect.Size())
	return 1 - (overlap / (float64(aColors.Size()) + float64(bColors.Size()) - overlap))
}

func (a *GraphPattern) CommonAncestor(o lattice.Pattern) lattice.Pattern {
	b := o.(*GraphPattern)
	if a.Level() > b.Level() {
		a, b = b, a
	}
	if a.Equals(b) {
		return a
	} else if a.Level() < 1 {
		return EmptyPattern
	} else if b.Level() < 1 {
		return EmptyPattern
	}
	aSet := set.FromSlice([]types.Hashable{a})
	bSet := set.FromSlice([]types.Hashable{b})
	firstB, _ := bSet.Get(0)
	// errors.Logf("DEBUG", "levels: a %v b %v", a.Level(), b.Level())
	for firstB.(*GraphPattern).Level() > a.Level() {
		nbSet := set.NewSortedSet(bSet.Size()+1)
		for x, next := bSet.Items()(); next != nil; x, next = next() {
			v := x.(*GraphPattern)
	// 		errors.Logf("DEBUG", "sg: %v", v.sg)
			parents := v.sg.SubGraphs()
			for _, p := range parents {
				nbSet.Add(NewGraphPattern(p))
			}
		}
		bSet = nbSet
		firstB, _ = bSet.Get(0)
	}
	// errors.Logf("DEBUG", "levels: a %v b %v len(b) %v", a.Level(), firstB.(*GraphPattern).Level(), bSet.Size())
	for !aSet.Overlap(bSet) {
		naSet := set.NewSortedSet(aSet.Size())
		nbSet := set.NewSortedSet(bSet.Size())
		for x, next := aSet.Items()(); next != nil; x, next = next() {
			v := x.(*GraphPattern)
			parents := v.sg.CanonSubGraphs()
			for _, p := range parents {
				naSet.Add(NewGraphPattern(p))
			}
		}
		for x, next := bSet.Items()(); next != nil; x, next = next() {
			v := x.(*GraphPattern)
			parents := v.sg.CanonSubGraphs()
			for _, p := range parents {
				nbSet.Add(NewGraphPattern(p))
			}
		}
		aSet = naSet
		bSet = nbSet
		if aSet.Size() == 0 || bSet.Size() == 0 {
			return EmptyPattern
		}
	}
	ancs, err := aSet.Intersect(bSet)
	if err != nil {
		log.Fatal(err)
	}
	anc, err := ancs.(*set.SortedSet).Random()
	if err != nil {
		log.Fatal(err)
	}
	return anc.(*GraphPattern)
}

func (g *GraphPattern) Equals(o types.Equatable) bool {
	a := types.ByteSlice(g.Label())
	switch b := o.(type) {
	case *GraphPattern: return a.Equals(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (g *GraphPattern) Less(o types.Sortable) bool {
	a := types.ByteSlice(g.Label())
	switch b := o.(type) {
	case *GraphPattern: return a.Less(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (g *GraphPattern) Hash() int {
	return types.ByteSlice(g.Label()).Hash()
}



func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
