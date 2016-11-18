package itemset

import (
	"encoding/binary"
	"fmt"
	"log"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/exc"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
)

type Pattern struct {
	Items *set.SortedSet
}

type Node struct {
	pat Pattern
	dt  *ItemSets
	txs []int32
}

type Embedding struct {
	tx int32
}

func setToInt32s(s *set.SortedSet) []int32 {
	items := make([]int32, 0, s.Size())
	for i, n := s.Items()(); n != nil; i, n = n() {
		item := int32(i.(types.Int32))
		items = append(items, item)
	}
	return items
}

func int32sToSet(list []int32) *set.SortedSet {
	items := set.NewSortedSet(len(list))
	for _, item := range list {
		items.Add(types.Int32(item))
	}
	return items
}

func TryLoadNode(items []int32, dt *ItemSets) (n *Node, _ error) {
	err := dt.Embeddings.DoFind(items, func(key, txs []int32) error {
		n = &Node{Pattern{int32sToSet(key)}, dt, txs}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return n, nil
}

func LoadNode(items []int32, dt *ItemSets) (n *Node, err error) {
	n, err = TryLoadNode(items, dt)
	if err != nil {
		return nil, err
	} else if n == nil {
		return nil, errors.Errorf("Expected node %v to be in embeddings store but it wasn't", items)
	}
	return n, nil
}

func (n *Node) Pattern() lattice.Pattern {
	return &n.pat
}

func (n *Node) Save() error {
	key := setToInt32s(n.pat.Items)
	if has, err := n.dt.Embeddings.Has(key); err != nil {
		return err
	} else if has {
		return nil
	}
	return n.dt.Embeddings.Add(key, n.txs)
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.pat.Items, len(n.txs))
}

func (n *Node) Parents() ([]lattice.Node, error) {
	if n.pat.Items.Size() == 0 {
		return []lattice.Node{}, nil
	} else if n.pat.Items.Size() == 1 {
		return []lattice.Node{n.dt.empty}, nil
	}
	i := setToInt32s(n.pat.Items)
	if has, err := n.dt.ParentCount.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(n.dt.Parents, i)
	}
	parents := make([]*set.SortedSet, 0, n.pat.Items.Size())
	for item, next := n.pat.Items.Items()(); next != nil; item, next = next() {
		parent := n.pat.Items.Copy()
		parent.Delete(item)
		parents = append(parents, parent)
	}
	nodes := make([]lattice.Node, 0, 10)
	for _, items := range parents {
		if node, err := TryLoadNode(setToInt32s(items), n.dt); err != nil {
			return nil, err
		} else if node != nil {
			nodes = append(nodes, node)
			continue
		}
		ctxs := int32sToSet(n.txs)
		var txs types.Set
		for item, next := items.Items()(); next != nil; item, next = next() {
			mytxs := set.NewSortedSet(len(n.txs) + 10)
			for _, tx := range n.dt.InvertedIndex[item.(types.Int32)] {
				if !ctxs.Has(types.Int32(tx)) {
					mytxs.Add(types.Int32(tx))
				}
			}
			var err error
			if txs == nil {
				txs = mytxs
			} else {
				txs, err = txs.Intersect(mytxs)
				if err != nil {
					return nil, err
				}
			}
		}
		txs, err := txs.Union(ctxs)
		if err != nil {
			return nil, err
		}
		stxs := make([]int32, 0, txs.Size())
		for item, next := txs.Items()(); next != nil; item, next = next() {
			stxs = append(stxs, int32(item.(types.Int32)))
		}
		node := &Node{Pattern{items}, n.dt, stxs}
		err = node.Save()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	err := n.cache(n.dt.ParentCount, n.dt.Parents, i, nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (n *Node) Children() ([]lattice.Node, error) {
	return n.kids(n.dt.ChildCount, n.dt.Children, n.allCandidateKids)
}

func (n *Node) CanonKids() ([]lattice.Node, error) {
	return n.kids(n.dt.CanonKidCount, n.dt.CanonKids, n.canonCandidateKids)
}

func (n *Node) kids(counts ints_int.MultiMap, kids ints_ints.MultiMap, candidates func() map[int32][]int32) ([]lattice.Node, error) {
	if n.pat.Items.Size() == 0 {
		return n.dt.FrequentItems, nil
	}
	if n.pat.Items.Size() >= n.dt.MaxItems {
		return []lattice.Node{}, nil
	}
	i := setToInt32s(n.pat.Items)
	if has, err := counts.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(kids, i)
	}
	exts := candidates()
	nodes, err := n.nodesFromCandidateKids(exts)
	if err != nil {
		return nil, err
	}
	err = n.cache(counts, kids, i, nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (n *Node) canonCandidateKids() map[int32][]int32 {
	// this works because n.pat.Items is a set.SortedSet
	l, err := n.pat.Items.Get(n.pat.Items.Size() - 1)
	if err != nil {
		log.Fatal(err)
	}
	largest := int32(l.(types.Int32))
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		for _, item := range n.dt.Index[tx] {
			if item <= largest {
				continue
			}
			if !n.pat.Items.Has(types.Int32(item)) {
				exts[item] = append(exts[item], tx)
			}
		}
	}
	return exts
}

func (n *Node) allCandidateKids() map[int32][]int32 {
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		for _, item := range n.dt.Index[tx] {
			if !n.pat.Items.Has(types.Int32(item)) {
				exts[item] = append(exts[item], tx)
			}
		}
	}
	return exts
}

func (n *Node) nodesFromCandidateKids(exts map[int32][]int32) ([]lattice.Node, error) {
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range exts {
		if len(txs) >= n.dt.Support() && !n.pat.Items.Has(types.Int32(item)) {
			items := n.pat.Items.Copy()
			items.Add(types.Int32(item))
			node := &Node{
				pat: Pattern{items},
				dt:  n.dt,
				txs: txs,
			}
			err := node.Save()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
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
	i := setToInt32s(n.pat.Items)
	if has, err := n.dt.ParentCount.Has(i); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Parents()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ParentCount.DoFind(i, func(_ []int32, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) ChildCount() (int, error) {
	i := setToInt32s(n.pat.Items)
	if has, err := n.dt.ChildCount.Has(i); err != nil {
		return 0, err
	} else if !has {
		nodes, err := n.Children()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := n.dt.ChildCount.DoFind(i, func(_ []int32, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) Maximal() (bool, error) {
	count, err := n.ChildCount()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (n *Node) cache(counts ints_int.MultiMap, m ints_ints.MultiMap, key []int32, nodes []lattice.Node) error {
	for _, node := range nodes {
		err := m.Add(key, setToInt32s(node.(*Node).pat.Items))
		if err != nil {
			return err
		}
	}
	return counts.Add(key, int32(len(nodes)))
}

func (n *Node) cached(m ints_ints.MultiMap, key []int32) (nodes []lattice.Node, _ error) {
	nodes = make([]lattice.Node, 0, 10)
	doerr := m.DoFind(key,
		func(_, value []int32) error {
			node, err := LoadNode(value, n.dt)
			if err != nil {
				return err
			}
			nodes = append(nodes, node)
			return nil
		})
	if doerr != nil {
		return nil, doerr
	}
	return nodes, nil
}

func (i *Pattern) Label() []byte {
	size := uint32(i.Items.Size())
	bytes := make([]byte, 4*(size+1))
	binary.BigEndian.PutUint32(bytes[0:4], size)
	s := 4
	e := s + 4
	for item, next := i.Items.Items()(); next != nil; item, next = next() {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(int32(item.(types.Int32))))
		s += 4
		e = s + 4
	}
	return bytes
}

func (i *Pattern) Level() int {
	return i.Items.Size() + 1
}

func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (a *Pattern) Distance(p lattice.Pattern) float64 {
	b := p.(*Pattern)
	i, err := a.Items.Intersect(b.Items)
	exc.ThrowOnError(err)
	inter := float64(i.Size())
	return 1.0 - (inter / (float64(a.Items.Size()) + float64(b.Items.Size()) - inter))
}



func (i *Pattern) Equals(o types.Equatable) bool {
	switch b := o.(type) {
	case *Pattern:
		return i.Items.Equals(b.Items)
	default:
		return false
	}
}

func (i *Pattern) Less(o types.Sortable) bool {
	switch b := o.(type) {
	case *Pattern:
		return i.Items.Less(b.Items)
	default:
		return false
	}
}

func (i *Pattern) Hash() int {
	return i.Items.Hash()
}
