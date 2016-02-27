package digraph

import (
	"bytes"
	"fmt"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

type Pattern struct {
	label []byte
	level int
	Sg *goiso.SubGraph
}

var EmptyPattern = &Pattern{
	label: []byte{},
	level: 0,
	Sg: nil,
}

type Node struct {
	pat     Pattern
	dt      *Graph
	sgs     SubGraphs
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

func NewNode(dt *Graph, label []byte, sgs SubGraphs) *Node {
	return &Node{Pattern{label: label}, dt, sgs}
}

func LoadNode(dt *Graph, label []byte) (*Node, error) {
	sgs := make(SubGraphs, 0, 10)
	err := dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewNode(dt, label, sgs), nil
}

func (n *Node) Pattern() lattice.Pattern {
	n.pat.level = n.Level()
	if n.pat.level > 0 {
		n.pat.Sg = n.sgs[0]
	}
	return &n.pat
}

func (n *Node) Save() error {
	if has, err := n.dt.Embeddings.Has(n.pat.label); err != nil {
		return err
	} else if has {
		return nil
	}
	for _, sg := range n.sgs {
		err := n.dt.Embeddings.Add(n.pat.label, sg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) String() string {
	if len(n.sgs) > 0 {
		return fmt.Sprintf("<Node %v>", n.sgs[0].Label())
	} else {
		return fmt.Sprintf("<Node {}>")
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
		return []lattice.Node{NewNode(n.dt, nil, nil)}, nil
	}
	if nodes, has, err := cached(n.dt, n.dt.ParentCount, n.dt.Parents, n.pat.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	parents := make([]lattice.Node, 0, 10)
	for _, parent := range n.sgs[0].SubGraphs() {
		p, err := FindNode(n.dt, parent)
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		} else if p != nil {
			parents = append(parents, p)
		}
	}
	if len(parents) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return parents, cache(n.dt, n.dt.ParentCount, n.dt.Parents, n.pat.label, parents)
}

func FindNode(dt *Graph, target *goiso.SubGraph) (*Node, error) {
	label := target.ShortLabel()
	if has, err := dt.Embeddings.Has(label); err != nil {
		return nil, err
	} else if has {
		return LoadNode(dt, label)
	}
	// errors.Logf("DEBUG", "target %v", target.Label())
	cur, graphs, err := edgeChain(dt, target)
	if err != nil {
		return nil, err
	}
	err = cur.sgs.Verify()
	if err != nil {
		return nil, err
	}
	for _, sg := range graphs {
		// errors.Logf("DEBUG", "extend %v to %v", cur, sg.Label())
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
		parents, err := allParents(cur)
		if err != nil {
			return nil, nil, err
		} else if len(parents) == 0 {
			return nil, nil, errors.Errorf("Could not find a connected parent!")
		} else {
			graphs[chainIdx] = cur
			cur = parents[0]
		}
	}
	// for i, g := range graphs {
	// errors.Logf("DEBUG", "graph %v %v", i, g.Label())
	// }
	startSg, _ := dt.G.VertexSubGraph(cur.V[0].Id)
	startLabel := startSg.ShortLabel()
	node, err := LoadNode(dt, startLabel)
	if err != nil {
		return nil, nil, err
	}
	return node, graphs, nil
}

func (n *Node) extendTo(sg *goiso.SubGraph) (exts *Node, err error) {
	// errors.Logf("DEBUG", "doing extendTo extend %v to %v", n, sg.Label())
	latKids, err := n.Children()
	if err != nil {
		return nil, err
	}
	label := sg.ShortLabel()
	for _, lkid := range latKids {
		kid := lkid.(*Node)
		// errors.Logf("DEBUG", "trying to find extension, have %v", kid)
		if bytes.Equal(label, kid.pat.label) {
			// errors.Logf("DEBUG", "FOUND extension\n    from %v\n      to %v\n    have %v", n.sgs[0].Label(), sg.Label(), kid.sgs[0].Label())
			return kid, nil
		}
	}
	// errors.Logf("DEBUG", "could not find extension\n    from %v\n      to %v", n.sgs[0].Label(), sg.Label())
	return nil, nil
}

func (n *Node) Children() (nodes []lattice.Node, err error) {
	return n.children(false, n.dt.Children, n.dt.ChildCount)
}

func (n *Node) CanonKids() (nodes []lattice.Node, err error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return n.children(true, n.dt.CanonKids, n.dt.CanonKidCount)
}

func (n *Node) children(checkCanon bool, children bytes_bytes.MultiMap, childCount bytes_int.MultiMap) (nodes []lattice.Node, err error) {
	if len(n.sgs) == 0 {
		return n.dt.FrequentVertices, nil
	}
	if len(n.sgs[0].E) >= n.dt.MaxEdges {
		return []lattice.Node{}, nil
	}
	if nodes, has, err := cached(n.dt, childCount, children, n.pat.label); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	// errors.Logf("DEBUG", "Children of %v", n)
	exts := make(SubGraphs, 0, 10)
	add := func(exts SubGraphs, sg *goiso.SubGraph, e *goiso.Edge) (SubGraphs, error) {
		if n.dt.G.ColorFrequency(e.Color) < n.dt.Support() {
			return exts, nil
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Src].Color) < n.dt.Support() {
			return exts, nil
		} else if n.dt.G.ColorFrequency(n.dt.G.V[e.Targ].Color) < n.dt.Support() {
			return exts, nil
		}
		if !sg.HasEdge(goiso.ColoredArc{e.Arc, e.Color}) {
			ext, _ := sg.EdgeExtend(e)
			if len(ext.V) > n.dt.MaxVertices {
				return exts, nil
			}
			exts = append(exts, ext)
		}
		return exts, nil
	}
	for _, sg := range n.sgs {
		for _, u := range sg.V {
			for _, e := range n.dt.G.Kids[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
			for _, e := range n.dt.G.Parents[u.Id] {
				exts, err = add(exts, sg, e)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	// errors.Logf("DEBUG", "len(exts) %v", len(exts))
	partitioned := exts.Partition()
	sum := 0
	for _, sgs := range partitioned {
		sum += len(sgs)
		// errors.Logf("DEBUG", "len(partition) %v %v", len(sgs), sgs[0].Label())
		if len(sgs) < n.dt.Support() {
			continue
		}
		supported := n.dt.Supported(sgs)
		// errors.Logf("DEBUG", "len(supported) %v %v", len(sgs), sgs[0].Label())
		if len(supported) >= n.dt.Support() {
			label := sgs[0].ShortLabel()
			if checkCanon {
				if canonized, err := isCanonicalExtension(n.sgs[0], sgs[0]); err != nil {
					return nil, err
				} else if !canonized {
					// errors.Logf("DEBUG", "%v is not canon (skipping)", sgs[0].Label())
				} else {
					nodes = append(nodes, NewNode(n.dt, label, sgs))
				}
			} else {
				nodes = append(nodes, NewNode(n.dt, label, sgs))
			}
		}
	}
	// errors.Logf("DEBUG", "sum(len(partition)) %v", sum)
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, cache(n.dt, childCount, children, n.pat.label, nodes)
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
	if has, err := n.dt.ParentCount.Has(n.pat.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Parents()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ParentCount.DoFind(n.pat.label, func(_ []byte, c int32) error {
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
	if has, err := n.dt.ChildCount.Has(n.pat.label); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Children()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ChildCount.DoFind(n.pat.label, func(_ []byte, c int32) error {
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

func NewPattern(sg *goiso.SubGraph) *Pattern {
	return &Pattern{
		label: sg.ShortLabel(),
		level: len(sg.E) + 1,
		Sg: sg,
	}
}

func (g *Pattern) Label() []byte {
	return g.label
}

func (g *Pattern) Level() int {
	return g.level
}

func (g *Pattern) String() string {
	if g.Sg != nil {
		return fmt.Sprintf("<Pattern %v>", g.Sg.Label())
	} else {
		return fmt.Sprintf("<Pattern {}>")
	}
}

func (g *Pattern) Equals(o types.Equatable) bool {
	a := types.ByteSlice(g.Label())
	switch b := o.(type) {
	case *Pattern: return a.Equals(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (g *Pattern) Less(o types.Sortable) bool {
	a := types.ByteSlice(g.Label())
	switch b := o.(type) {
	case *Pattern: return a.Less(types.ByteSlice(b.Label()))
	default: return false
	}
}

func (g *Pattern) Hash() int {
	return types.ByteSlice(g.Label()).Hash()
}



func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
