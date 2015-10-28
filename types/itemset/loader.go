package itemset

import (
	"io"
	"bufio"
	"strconv"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/intint"
)


type MakeLoader func(*ItemSets) lattice.Loader


type ItemSets struct {
	InvertedIndex intint.MultiMap
	FrequentItems []*Node
	makeLoader MakeLoader
	config *config.Config
}

func NewItemSets(config *config.Config, makeLoader MakeLoader) (i *ItemSets, err error) {
	var index intint.MultiMap
	if config.Cache == "" {
		index, err = intint.AnonBpTree()
	} else {
		index, err = intint.NewBpTree(config.CacheFile("itemsets-inverted.bptree"))
	}
	if err != nil {
		return nil, err
	}
	i = &ItemSets{
		InvertedIndex: index,
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
	return i.InvertedIndex.Delete()
}


type IntLoader struct {
	sets *ItemSets
}

func NewIntLoader(sets *ItemSets) lattice.Loader {
	return &IntLoader{
		sets: sets,
	}
}

func (l *IntLoader) StartingPoints(input io.Reader, support int) ([]lattice.Node, error) {
	scanner := bufio.NewScanner(input)
	tx := 0
	for scanner.Scan() {
		line := scanner.Text()
		for _, col := range strings.Split(line, " ") {
			if col == "" {
				continue
			}
			item, err := strconv.Atoi(col)
			if err != nil {
				errors.Logf("WARN", "input line %d contained non int '%s'", tx, col)
			}
			err = l.sets.InvertedIndex.Add(item, tx)
			if err != nil{
				return nil, err
			}
		}
		tx += 1
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	nodes := make([]lattice.Node, 0, 10)
	citem := -1
	txs := make([]int, 0, 10)
	err := intint.Do(l.sets.InvertedIndex.Iterate, func(item, tx int) error {
		if len(txs) > 0 && item != citem {
			n := &Node{
				items: []int{citem},
				embeddings: txs,
			}
			l.sets.FrequentItems = append(l.sets.FrequentItems, n)
			nodes = append(nodes, n)
			txs = make([]int, 0, 10)
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
			items: []int{citem},
			embeddings: txs,
		}
		l.sets.FrequentItems = append(l.sets.FrequentItems, n)
		nodes = append(nodes, n)
	}
	return nodes, nil
}

