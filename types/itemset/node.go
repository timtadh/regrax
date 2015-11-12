package itemset

import (
	"encoding/binary"
	"fmt"
)

import (
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/intint"
	"github.com/timtadh/sfp/stores/itemset_int"
	"github.com/timtadh/sfp/stores/itemsets"
)


type Node struct {
	items *set.SortedSet
	txs []int32
}

type Embedding struct {
	tx int32
}

func (n *Node) ItemSet() *itemsets.ItemSet {
	items := make([]int32, 0, n.items.Size())
	for i, n := n.items.Items()(); n != nil; i, n = n() {
		item := int32(i.(types.Int32))
		items = append(items, item)
	}
	txs := make([]int32, len(n.txs))
	copy(txs, n.txs)
	return &itemsets.ItemSet{
		Items: items,
		Txs: txs,
	}
}

func NodeFromItemSet(i *itemsets.ItemSet) *Node {
	items := set.NewSortedSet(len(i.Items))
	for _, item := range i.Items {
		items.Add(types.Int32(item))
	}
	txs := make([]int32, len(i.Txs))
	copy(txs, i.Txs)
	return &Node {
		items: items,
		txs: txs,
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.items, len(n.txs))
}

func (n *Node) StartingPoint() bool {
	return n.Size() == 1
}

func (n *Node) Size() int {
	return n.items.Size()
}

func (n *Node) Parents(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	if n.items.Size() == 1 {
		return []lattice.Node{}, nil
	}
	dt := dtype.(*ItemSets)
	i := n.ItemSet()
	if has, err := dt.ParentCount.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(dt.Parents.Find(i))
	}
	parents := make([]*set.SortedSet, 0, n.items.Size())
	for item, next := n.items.Items()(); next != nil; item, next = next() {
		parent := n.items.Copy()
		parent.Delete(item)
		parents = append(parents, parent)
	}
	nodes := make([]lattice.Node, 0, 10)
	for _, items := range parents {
		var txs types.Set
		for item, next := items.Items()(); next != nil; item, next = next() {
			mytxs := set.NewSortedSet(len(n.txs)+10)
			err := intint.Do(
				func() (intint.Iterator, error) {
					return dt.InvertedIndex.Find(int32(item.(types.Int32)))
				},
				func(item, tx int32) error {
					return mytxs.Add(types.Int32(tx))
				})
			if err != nil {
				return nil, err
			}
			if txs == nil {
				txs = mytxs
			} else {
				txs, err = txs.Intersect(mytxs)
				if err != nil {
					return nil, err
				}
			}
		}
		stxs := make([]int32, 0, txs.Size())
		for item, next := txs.Items()(); next != nil; item, next = next() {
			stxs = append(stxs, int32(item.(types.Int32)))
		}
		nodes = append(nodes, &Node{items, stxs})
	}
	err := n.cache(dt.ParentCount, dt.Parents, i, nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (n *Node) Children(support int, dtype lattice.DataType) ([]lattice.Node, error) {
	dt := dtype.(*ItemSets)
	i := n.ItemSet()
	if has, err := dt.ChildCount.Has(i); err != nil {
		return nil, err
	} else if has {
		return n.cached(dt.Children.Find(i))
	}
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		err := dt.Index.DoFind(tx,
			func(tx, item int32) error {
				if !n.items.Has(types.Int32(item)) {
					exts[item] = append(exts[item], tx)
				}
				return nil
			})
		if err != nil {
			return nil, err
		}
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range exts {
		if len(txs) >= support {
			items := n.items.Copy()
			items.Add(types.Int32(item))
			n := &Node{
				items: items,
				txs: txs,
			}
			nodes = append(nodes, n)
		}
	}
	err := n.cache(dt.ChildCount, dt.Children, i, nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (n *Node) ChildCount(support int, dtype lattice.DataType) (int, error) {
	dt := dtype.(*ItemSets)
	i := n.ItemSet()
	var count int32
	err := dt.ChildCount.DoFind(i, func(_ *itemsets.ItemSet, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (n *Node) Maximal(support int, dtype lattice.DataType) (bool, error) {
	count, err := n.ChildCount(support, dtype)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (n *Node) cache(counts itemset_int.MultiMap, m itemsets.MultiMap, key *itemsets.ItemSet, nodes []lattice.Node) error {
	for _, node := range nodes {
		err := m.Add(key, node.(*Node).ItemSet())
		if err != nil {
			return err
		}
	}
	return counts.Add(key, int32(len(nodes)))
}

func (n *Node) cached(it itemsets.Iterator, err error) (nodes []lattice.Node, _ error) {
	nodes = make([]lattice.Node, 0, 10)
	doerr := itemsets.Do(
		func()(itemsets.Iterator, error) { return it, err },
		func(key, value *itemsets.ItemSet) error {
			nodes = append(nodes, NodeFromItemSet(value))
			return nil
		})
	if doerr != nil {
		return nil, doerr
	}
	return nodes, nil
}

func (n *Node) Label() []byte {
	size := uint32(n.items.Size())
	bytes := make([]byte, 4*(size + 1))
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
		embeddings = append(embeddings, &Embedding{tx:tx})
	}
	return embeddings, nil
}

func (n *Node) Lattice(support int, dtype lattice.DataType) (*lattice.Lattice, error) {
	return nil, &lattice.NoLattice{}
}

func (e *Embedding) Components() ([]int, error) {
	return []int{int(e.tx)}, nil
}

