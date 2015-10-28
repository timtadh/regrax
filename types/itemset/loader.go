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
)


type MakeLoader func(*ItemSets) lattice.Loader


type ItemSets struct {
	InvertedIndex map[int][]int
	FrequentItems []*Node
	makeLoader MakeLoader
	config *config.Config
}

func NewItemSets(config *config.Config, makeLoader MakeLoader) *ItemSets {
	return &ItemSets{
		InvertedIndex: make(map[int][]int),
		FrequentItems: make([]*Node, 0, 10),
		makeLoader: makeLoader,
		config: config,
	}
}

func (i *ItemSets) Metric() lattice.SupportMetric {
	return lattice.RawSupport{}
}

func (i *ItemSets) Loader() lattice.Loader {
	return i.makeLoader(i)
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
			l.addInverted(item, tx)
		}
		tx += 1
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	nodes := make([]lattice.Node, 0, 10)
	for item, txs := range l.sets.InvertedIndex {
		if len(txs) >= support {
			n := &Node{
				items: []int{item},
				embeddings: txs,
			}
			l.sets.FrequentItems = append(l.sets.FrequentItems, n)
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (l *IntLoader) addInverted(item, tx int) error {
	if txs := l.sets.InvertedIndex[item]; txs == nil {
		l.sets.InvertedIndex[item] = make([]int, 0, 10)
	}
	l.sets.InvertedIndex[item] = append(l.sets.InvertedIndex[item], tx)
	return nil
}

