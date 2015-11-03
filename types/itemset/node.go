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
)


type Node struct {
	items *set.SortedSet
	txs []int32
}

type Embedding struct {
	tx int32
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.items, len(n.txs))
}

func (n *Node) Size() int {
	return n.items.Size()
}

func (n *Node) Parents(support int, dtype lattice.DataType) (lattice.NodeIterator, error) {
	dt := dtype.(*ItemSets)
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
	return lattice.NodeIteratorFromSlice(nodes)
}

func (n *Node) Children(support int, dtype lattice.DataType) (it lattice.NodeIterator, err error) {
	dt := dtype.(*ItemSets)
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		err := intint.Do(func() (intint.Iterator, error) {return dt.Index.Find(tx)},
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
	return lattice.NodeIteratorFromSlice(nodes)
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

func (e *Embedding) Components() ([]int, error) {
	return []int{int(e.tx)}, nil
}

