package itemset

import (
	"io"
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
		invIndex, err = intint.AnonBpTree()
	} else {
		index, err = intint.NewBpTree(config.CacheFile("itemsets-index.bptree"))
		invIndex, err = intint.NewBpTree(config.CacheFile("itemsets-inverted.bptree"))
	}
	if err != nil {
		return nil, err
	}
	i = &ItemSets{
		Index: index,
		InvertedIndex: invIndex,
		FrequentItems: make([]*Node, 0, 10),
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
	i.InvertedIndex.Close()
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

func (l *IntLoader) buildIndex(input io.Reader, support int) (error) {
	scanner := bufio.NewScanner(input)
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
			/*
			err = l.sets.Index.Add(tx, item)
			if err != nil{
				return err
			}
			*/
			err = l.sets.InvertedIndex.Add(int32(item), tx)
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

func (l *IntLoader) StartingPoints(input io.Reader, support int) ([]lattice.Node, error) {
	err := l.buildIndex(input, support)
	if err != nil {
		return nil, err
	}
	nodes := make([]lattice.Node, 0, 10)
	citem := int32(-1)
	txs := make([]int32, 0, 10)
	err = intint.Do(l.sets.InvertedIndex.Iterate, func(item, tx int32) error {
		if len(txs) > 0 && item != citem {
			n := &Node{
				items: set.FromSlice([]types.Hashable{types.Int32(citem)}),
				embeddings: txs,
			}
			l.sets.FrequentItems = append(l.sets.FrequentItems, n)
			nodes = append(nodes, n)
			txs = make([]int32, 0, 10)
		}
		citem = item
		txs = append(txs, tx)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(txs) > 0 {
		n := &Node{
			items: set.FromSlice([]types.Hashable{types.Int32(citem)}),
			embeddings: txs,
		}
		l.sets.FrequentItems = append(l.sets.FrequentItems, n)
		nodes = append(nodes, n)
	}
	return nodes, nil
}

