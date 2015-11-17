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
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
)


type MakeLoader func(*ItemSets) lattice.Loader
type itemsIter func(func(tx, item int32) error) error


type ItemSets struct {
	Index int_int.MultiMap
	InvertedIndex int_int.MultiMap
	Parents ints_ints.MultiMap
	ParentCount ints_int.MultiMap
	Children ints_ints.MultiMap
	ChildCount ints_int.MultiMap
	Embeddings ints_ints.MultiMap
	makeLoader MakeLoader
	config *config.Config
}

func NewItemSets(config *config.Config, makeLoader MakeLoader) (i *ItemSets, err error) {
	index, err := config.IntIntMultiMap("itemsets-index")
	if err != nil {
		return nil, err
	}
	invIndex, err := config.IntIntMultiMap("itemsets-inv-index")
	if err != nil {
		return nil, err
	}
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
		Index: index,
		InvertedIndex: invIndex,
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
	i.Index.Close()
	i.InvertedIndex.Close()
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

func (l *IntLoader) maxItem(items itemsIter) (int32, error) {
	max := int32(0)
	err := items(func (tx, item int32) error {
		if item > max {
			max = item
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return max, nil
}

func (l *IntLoader) invert(items itemsIter) ([][]int32, error) {
	max, err := l.maxItem(items)
	if err != nil {
		return nil, err
	}
	inverted := make([][]int32, max+1)
	err = items(func (tx, item int32) error {
		if item >= int32(len(inverted)) {
			errors.Logf("DEBUG", "item = %v, max = %v", item, max)
		}
		inverted[item] = append(inverted[item], tx)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inverted, nil
}

func (l *IntLoader) buildIndex(input lattice.Input, inverted [][]int, support int) (error) {
	return nil
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
	inverted, err := l.invert(items)
	if err != nil {
		return nil, err
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range inverted {
		if len(txs) >= support {
			errors.Logf("INFO", "item %d len(txs) %d", item, len(txs))
			for _, tx := range txs {
				err := l.sets.Index.Add(tx, int32(item))
				if err != nil {
					return nil, err
				}
				err = l.sets.InvertedIndex.Add(int32(item), tx)
				if err != nil {
					return nil, err
				}
			}
			n := &Node{
				items: set.FromSlice([]types.Hashable{types.Int32(item)}),
				txs: txs,
			}
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

