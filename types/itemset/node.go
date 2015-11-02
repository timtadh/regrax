package itemset

import (
	"fmt"
)

import (
	"github.com/timtadh/data-structures/errors"
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

func (n *Node) Parents(support int, dtype lattice.DataType) (lattice.NodeIterator, error) {
	panic("unimplemented")
}

func (n *Node) String() string {
	return fmt.Sprintf("<Node %v %v>", n.items, len(n.txs))
}

func (n *Node) Children(support int, dtype lattice.DataType) (it lattice.NodeIterator, err error) {
	dt := dtype.(*ItemSets)
	errors.Logf("INFO", "dt %v", dt)
	exts := make(map[int32][]int32)
	for _, tx := range n.txs {
		intint.Do(func() (intint.Iterator, error) {return dt.Index.Find(tx)},
			func(tx, item int32) error {
				if !n.items.Has(types.Int32(item)) {
					exts[item] = append(exts[item], tx)
				}
				return nil
			})
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range exts {
		if len(txs) >= support {
			items := set.NewSortedSet(n.items.Size()+1)
			items.Extend(n.items.Items())
			items.Add(types.Int32(item))
			n := &Node{
				items: items,
				txs: txs,
			}
			nodes = append(nodes, n)
		}
	}
	i := 0
	it = func() (lattice.Node, error, lattice.NodeIterator) {
		if i >= len(nodes) {
			return nil, nil, nil
		}
		n := nodes[i]
		i++
		return n, nil, it
	}
	return it, nil
}

func (n *Node) Label() ([]byte, error) {
	panic("unimplemented")
}

func (n *Node) Embeddings() ([]lattice.Embedding, error) {
	panic("unimplemented")
}

