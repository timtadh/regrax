package itemset

import "testing"
import "github.com/stretchr/testify/assert"

import (
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
)


var items = [][]int32{
	{0},
	{1,2,3},
	{1,2,3},
	{1,2,3},
	{2,3,4},
	{2,3,4},
	{2,3,4},
	{7,8,9,10},
	{7,8,9,11},
	{7,8,9,12},
	{1,12},
	{1,11},
	{1,10},
	{1,8,10},
	{1,9,11},
	{1,4,12},
	{1,12,7},
	{1,11,8},
	{1,10,12},
}

func iterItems(items [][]int32) itemsIter {
	return func(do func(tx, item int32) error) error {
		for tx, set := range items {
			for _, item := range set {
				err := do(int32(tx), item)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func startingPoints(t *assert.Assertions) ([]*Node, *ItemSets, int) {
	c := &config.Config{}
	i, err := NewItemSets(c, NewIntLoader)
	t.Nil(err)
	N, err := i.Loader().(*IntLoader).startingPoints(iterItems(items), 3)
	t.Nil(err)
	nodes := make([]*Node, 0, len(N))
	for _, node := range N {
		n := node.(*Node)
		nodes = append(nodes, n)
		t.True(len(n.txs) >= 3, "len(n.txs) %d < 3", len(n.txs))
	}
	return nodes, i, 3
}

func TestLoad(x *testing.T) {
	t := assert.New(x)
	startingPoints(t)
}

func TestKids_1(x *testing.T) {
	t := assert.New(x)
	nodes, dt, sup := startingPoints(t)
	n1 := nodes[0]
	kids, err := n1.Children(sup, dt)
	t.Nil(err)
	expected := set.FromSlice([]types.Hashable{
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(2)}),
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(3)}),
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(10)}),
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(11)}),
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(12)}),
	})
	var next *Node = nil
	for _, kid := range kids {
		if kid.(*Node).items.Has(types.Int32(2)) {
			next = kid.(*Node)
		}
		has := expected.Has(kid.(*Node).items)
		t.True(has, "%v not in %v", kid.(*Node).items, expected)
	}
	kids, err = next.Children(sup, dt)
	t.Nil(err)
	t.True(len(kids) == 1, "len(kids) %d != 1", len(kids))
	t.True(kids[0].(*Node).items.Equals(
	       set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(2),
	                                      types.Int32(3)})))
}

func TestParents_123(x *testing.T) {
	t := assert.New(x)
	_, dt, sup := startingPoints(t)
	n123 := &Node{
		items: set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(2), types.Int32(3)}),
		txs: []int32{1,2,3},
	}
	parents, err := n123.Parents(sup, dt)
	t.Nil(err, "%v", err)
	expected := set.FromSlice([]types.Hashable{
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(2)}),
		set.FromSlice([]types.Hashable{types.Int32(1), types.Int32(3)}),
		set.FromSlice([]types.Hashable{types.Int32(2), types.Int32(3)}),
	})
	for _, p := range parents {
		has := expected.Has(p.(*Node).items)
		t.True(has, "%v not in %v", p.(*Node).items, expected)
	}
}

