package itemset

import (
	"bufio"
	"strconv"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
)


type MakeLoader func(*ItemSets) lattice.Loader
type itemsIter func(func(tx, item int32) error) error

type index [][]int32

type ItemSets struct {
	Index index
	InvertedIndex index
	Parents ints_ints.MultiMap
	ParentCount ints_int.MultiMap
	Children ints_ints.MultiMap
	ChildCount ints_int.MultiMap
	Embeddings ints_ints.MultiMap
	FrequentItems []lattice.Node
	Empty lattice.Node
	makeLoader MakeLoader
	config *config.Config
}

func (i index) grow(size int32) index {
	newcap := len(i) + 1
	for newcap - 1 <= int(size) {
		if len(i) < 10000 {
			newcap *= 2
		} else {
			newcap += 100
		}
		errors.Logf("DEBUG", "expanding. len(i) %v newcap %v size %v", len(i), newcap, size)
	}
	n := make(index, newcap)
	copy(n, i)
	return n
}

func (i index) add(idx, value int32) index {
	if int(idx) >= len(i) {
		i = i.grow(idx)
	}
	i[idx] = append(i[idx], value)
	return i
}

func NewItemSets(config *config.Config, makeLoader MakeLoader) (i *ItemSets, err error) {
	parents, err := config.IntsIntsMultiMap("itemsets-parents")
	if err != nil {
		return nil, err
	}
	children, err := config.IntsIntsMultiMap("itemsets-children")
	if err != nil {
		return nil, err
	}
	childCount, err := config.IntsIntMultiMap("itemsets-child-count")
	if err != nil {
		return nil, err
	}
	parentCount, err := config.IntsIntMultiMap("itemsets-parent-count")
	if err != nil {
		return nil, err
	}
	embeddings, err := config.IntsIntsMultiMap("itemsets-embeddings")
	if err != nil {
		return nil, err
	}
	i = &ItemSets{
		Parents: parents,
		ParentCount: parentCount,
		Children: children,
		ChildCount: childCount,
		Embeddings: embeddings,
		makeLoader: makeLoader,
		config: config,
	}
	return i, nil
}

func (i *ItemSets) Loader() lattice.Loader {
	return i.makeLoader(i)
}

func (i *ItemSets) Close() error {
	i.Parents.Close()
	i.ParentCount.Close()
	i.Children.Close()
	i.ChildCount.Close()
	i.Embeddings.Close()
	return nil
}


type IntLoader struct {
	sets *ItemSets
}

func NewIntLoader(sets *ItemSets) lattice.Loader {
	return &IntLoader{
		sets: sets,
	}
}

func (l *IntLoader) max(items itemsIter) (max_tx, max_item int32, err error) {
	err = items(func (tx, item int32) error {
		max_tx++
		if item > max_item {
			max_item = item
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return max_tx, max_item, nil
}

func (l *IntLoader) indices(items itemsIter, support int) (idx, inv index, err error) {
	max_tx, max_item, err := l.max(items)
	if err != nil {
		return nil, nil, err
	}
	counts := make([]int, max_item + 1)
	err = items(func (tx, item int32) error {
		counts[item]++
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	errors.Logf("DEBUG", "max tx : %v, max item : %v", max_tx, max_item)
	idx = make(index, max_tx + 1)
	inv = make(index, max_item + 1)
	err = items(func (tx, item int32) error {
		if counts[item] > support {
			idx = idx.add(tx, item)
			inv = inv.add(item, tx)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return idx, inv, nil
}

func (l *IntLoader) items(input lattice.Input) func(do func(tx, item int32) error) error {
	return func(do func(tx, item int32) error) error {
		in, closer := input()
		defer closer()
		scanner := bufio.NewScanner(in)
		tx := int32(0)
		for scanner.Scan() {
			if tx % 1000 == 0 {
				errors.Logf("INFO", "line %d", tx)
			}
			line := scanner.Text()
			for _, col := range strings.Split(line, " ") {
				if col == "" {
					continue
				}
				item, err := strconv.Atoi(col)
				if err != nil {
					errors.Logf("WARN", "input line %d contained non int '%s'", tx, col)
					continue
				}
				err = do(tx, int32(item))
				if err != nil {
					errors.Logf("ERROR", "%v", err)
					return err
				}
			}
			tx += 1
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		return nil
	}
}

func (l *IntLoader) StartingPoints(input lattice.Input, support int) ([]lattice.Node, error) {
	return l.startingPoints(l.items(input), support)
}

func (l *IntLoader) startingPoints(items itemsIter, support int) ([]lattice.Node, error) {
	idx, inv, err := l.indices(items, support)
	if err != nil {
		return nil, err
	}
	l.sets.Index = idx
	l.sets.InvertedIndex = inv
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range inv {
		if len(txs) >= support {
			errors.Logf("INFO", "item %d len(txs) %d", item, len(txs))
			n := &Node{
				items: set.FromSlice([]types.Hashable{types.Int32(item)}),
				txs: txs,
			}
			nodes = append(nodes, n)
		}
	}
	l.sets.FrequentItems = nodes
	l.sets.Empty = &Node{int32sToSet([]int32{}), []int32{}}
	return nodes, nil
}

