package digraph

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

type EmbListNode struct {
	SearchNode
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

func NewEmbListNode(Dt *Digraph, sgs SubGraphs) *EmbListNode {
	if len(sgs) > 0 {
		return &EmbListNode{newSearchNode(Dt, sgs[0]), sgs}
	}
	return &EmbListNode{newSearchNode(Dt, nil), nil}
}

func (n *EmbListNode) New(sgs []*goiso.SubGraph) Node {
	return NewEmbListNode(n.Dt, sgs)
}

func LoadEmbListNode(Dt *Digraph, label []byte) (*EmbListNode, error) {
	sgs := make(SubGraphs, 0, 10)
	err := Dt.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
		sgs = append(sgs, sg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewEmbListNode(Dt, sgs), nil
}

func (n *EmbListNode) Pattern() lattice.Pattern {
	return &n.SearchNode
}

func (n *EmbListNode) Embeddings() ([]*goiso.SubGraph, error) {
	return n.sgs, nil
}

func (n *EmbListNode) Save() error {
	if has, err := n.Dt.Embeddings.Has(n.Label()); err != nil {
		return err
	} else if has {
		return nil
	}
	for _, sg := range n.sgs {
		err := n.Dt.Embeddings.Add(n.Label(), sg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *EmbListNode) String() string {
	if len(n.sgs) > 0 {
		return fmt.Sprintf("<EmbListNode %v>", n.sgs[0].Label())
	} else {
		return fmt.Sprintf("<EmbListNode {}>")
	}
}

func (n *EmbListNode) Parents() ([]lattice.Node, error) {
	// errors.Logf("DEBUG", "compute Parents\n    of %v", n)
	if len(n.sgs) == 0 {
		return []lattice.Node{}, nil
	}
	if len(n.sgs[0].V) == 1 && len(n.sgs[0].E) == 0 {
		return []lattice.Node{NewEmbListNode(n.Dt, nil)}, nil
	}
	if nodes, has, err := cached(n.Dt, n.Dt.ParentCount, n.Dt.Parents, n.Label()); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	parents := make([]lattice.Node, 0, 10)
	for _, parent := range n.sgs[0].SubGraphs() {
		p, err := FindEmbListNode(n.Dt, parent)
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		} else if p != nil {
			parents = append(parents, p)
		}
	}
	if len(parents) == 0 {
		return nil, errors.Errorf("Found no parents!!\n    node %v", n)
	}
	return parents, cache(n.Dt, n.Dt.ParentCount, n.Dt.Parents, n.Label(), parents)
}

func FindEmbListNode(Dt *Digraph, target *goiso.SubGraph) (*EmbListNode, error) {
	label := target.ShortLabel()
	if has, err := Dt.Embeddings.Has(label); err != nil {
		return nil, err
	} else if has {
		return LoadEmbListNode(Dt, label)
	}
	// errors.Logf("DEBUG", "target %v", target.Label())
	cur, graphs, err := edgeChain(Dt, target)
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

func edgeChain(Dt *Digraph, target *goiso.SubGraph) (start *EmbListNode, graphs []*goiso.SubGraph, err error) {
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
	startSg, _ := Dt.G.VertexSubGraph(cur.V[0].Id)
	startLabel := startSg.ShortLabel()
	node, err := LoadEmbListNode(Dt, startLabel)
	if err != nil {
		return nil, nil, err
	}
	return node, graphs, nil
}

func (n *EmbListNode) extendTo(sg *goiso.SubGraph) (exts *EmbListNode, err error) {
	// errors.Logf("DEBUG", "doing extendTo extend %v to %v", n, sg.Label())
	latKids, err := n.Children()
	if err != nil {
		return nil, err
	}
	label := sg.ShortLabel()
	for _, lkid := range latKids {
		kid := lkid.(*EmbListNode)
		// errors.Logf("DEBUG", "trying to find extension, have %v", kid)
		if bytes.Equal(label, kid.Label()) {
			// errors.Logf("DEBUG", "FOUND extension\n    from %v\n      to %v\n    have %v", n.sgs[0].Label(), sg.Label(), kid.sgs[0].Label())
			return kid, nil
		}
	}
	// errors.Logf("DEBUG", "could not find extension\n    from %v\n      to %v", n.sgs[0].Label(), sg.Label())
	return nil, nil
}

func (n *EmbListNode) Children() (nodes []lattice.Node, err error) {
	return children(n, false, n.Dt.Children, n.Dt.ChildCount)
}

func (n *EmbListNode) CanonKids() (nodes []lattice.Node, err error) {
	// errors.Logf("DEBUG", "CanonKids of %v", n)
	return children(n, true, n.Dt.CanonKids, n.Dt.CanonKidCount)
}

func (n *EmbListNode) loadFrequentVertices() ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, len(n.Dt.FrequentVertices))
	for _, label := range n.Dt.FrequentVertices {
		node, err := LoadEmbListNode(n.Dt, label)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (n *EmbListNode) AdjacentCount() (int, error) {
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

func (n *EmbListNode) ParentCount() (int, error) {
	if len(n.sgs) == 0 {
		return 0, nil
	}
	if has, err := n.Dt.ParentCount.Has(n.Label()); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Parents()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.Dt.ParentCount.DoFind(n.Label(), func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *EmbListNode) ChildCount() (int, error) {
	if len(n.sgs) == 0 {
		return len(n.Dt.FrequentVertices), nil
	}
	if has, err := n.Dt.ChildCount.Has(n.Label()); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Children()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.Dt.ChildCount.DoFind(n.Label(), func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *EmbListNode) Maximal() (bool, error) {
	cc, err := n.ChildCount()
	if err != nil {
		return false, err
	}
	return cc == 0, nil
}

func (n *EmbListNode) Label() []byte {
	return n.SearchNode.Label()
}

func (n *EmbListNode) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}
