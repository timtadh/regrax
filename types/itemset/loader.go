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
	"github.com/timtadh/sfp/stores/intint"
)


type MakeLoader func(*ItemSets) lattice.Loader


type ItemSets struct {
	Index intint.MultiMap
	InvertedIndex intint.MultiMap
	FrequentItems []*Node
	makeLoader MakeLoader
	config *config.Config
}

func NewItemSets(config *config.Config, makeLoader MakeLoader) (i *ItemSets, err error) {
	var index intint.MultiMap
	var invIndex intint.MultiMap
	if config.Cache == "" {
		index, err = intint.AnonBpTree()
	} else {
		index, err = intint.NewBpTree(config.CacheFile("itemsets-index.bptree"))
	}
	if err != nil {
		return nil, err
	}
	if config.Cache == "" {
		invIndex, err = intint.AnonBpTree()
	} else {
		invIndex, err = intint.NewBpTree(config.CacheFile("itemsets-inv-index.bptree"))
	}
	if err != nil {
		return nil, err
	}
	i = &ItemSets{
		Index: index,
		InvertedIndex: invIndex,
		makeLoader: makeLoader,
		config: config,
	}
	return i, nil
}

func (i *ItemSets) Metric() lattice.SupportMetric {
	return lattice.RawSupport{}
}

func (i *ItemSets) Loader() lattice.Loader {
	return i.makeLoader(i)
}

func (i *ItemSets) Close() error {
	i.Index.Close()
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

func (l *IntLoader) maxItem(input lattice.Input) (int32, error) {
	max := int32(0)
	err := l.items(input, func (tx, item int32) error {
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

func (l *IntLoader) invert(input lattice.Input) ([][]int32, error) {
	max, err := l.maxItem(input)
	if err != nil {
		return nil, err
	}
	inverted := make([][]int32, max+1)
	err = l.items(input, func (tx, item int32) error {
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

func (l *IntLoader) items(input lattice.Input, do func(tx, item int32) error) error {
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

func (l *IntLoader) StartingPoints(input lattice.Input, support int) ([]lattice.Node, error) {
	inverted, err := l.invert(input)
	if err != nil {
		return nil, err
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range inverted {
		errors.Logf("INFO", "item %d %d", item, len(txs))
		if len(txs) >= support {
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

