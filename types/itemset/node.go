package itemset

import (
	"encoding/binary"
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
)

type Node struct {
	dt    *ItemSets
	items *set.SortedSet
	txs   []int32
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
		n = &Node{dt, int32sToSet(key), txs}
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

func (n *Node) Save() error {
	key := setToInt32s(n.items)
	if has, err := n.dt.Embeddings.Has(key); err != nil {
		return err
	} else if has {
		return nil
	}
	return n.dt.Embeddings.Add(key, n.txs)
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.items, len(n.txs))
}

func (n *Node) Level() int {
	return n.items.Size() + 1
}

func (n *Node) Parents() ([]lattice.Node, error) {
	if n.items.Size() == 0 {
		return []lattice.Node{}, nil
	} else if n.items.Size() == 1 {
		return []lattice.Node{n.dt.empty}, nil
	}
	i := setToInt32s(n.items)
	if has, err := n.dt.ParentCount.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(n.dt.Parents, i)
	}
	parents := make([]*set.SortedSet, 0, n.items.Size())
	for item, next := n.items.Items()(); next != nil; item, next = next() {
		parent := n.items.Copy()
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
		node := &Node{n.dt, items, stxs}
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
	if n.items.Size() == 0 {
		return n.dt.FrequentItems, nil
	}
	if n.items.Size() >= n.dt.MaxItems {
		return []lattice.Node{}, nil
	}
	i := setToInt32s(n.items)
	if has, err := n.dt.ChildCount.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(n.dt.Children, i)
	}
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		for _, item := range n.dt.Index[tx] {
			if !n.items.Has(types.Int32(item)) {
				exts[item] = append(exts[item], tx)
			}
		}
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range exts {
		if len(txs) >= n.dt.Support() && !n.items.Has(types.Int32(item)) {
			items := n.items.Copy()
			items.Add(types.Int32(item))
			node := &Node{
				dt:    n.dt,
				items: items,
				txs:   txs,
			}
			err := node.Save()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
	err := n.cache(n.dt.ChildCount, n.dt.Children, i, nodes)
	if err != nil {
		return nil, err
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
	i := setToInt32s(n.items)
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
	i := setToInt32s(n.items)
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
		err := m.Add(key, setToInt32s(node.(*Node).items))
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

func (n *Node) Label() []byte {
	size := uint32(n.items.Size())
	bytes := make([]byte, 4*(size+1))
	binary.BigEndian.PutUint32(bytes[0:4], size)
	s := 4
	e := s + 4
	for item, next := n.items.Items()(); next != nil; item, next = next() {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(int32(item.(types.Int32))))
		s += 4
		e = s + 4
	}
	return bytes
}

func (n *Node) Embeddings() ([]lattice.Embedding, error) {
	embeddings := make([]lattice.Embedding, 0, len(n.txs))
	for _, tx := range n.txs {
		embeddings = append(embeddings, &Embedding{tx: tx})
	}
	return embeddings, nil
}

func (n *Node) Lattice() (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (e *Embedding) Components() ([]int, error) {
	return []int{int(e.tx)}, nil
}
