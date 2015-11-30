package itemset

import (
	"fmt"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Formatter struct{}

func (f Formatter) FileExt() string {
	return ".items"
}

func (f Formatter) PatternName(n lattice.Node) string {
	return f.FormatPattern(n)
}

func (f Formatter) FormatPattern(node lattice.Node) string {
	n := node.(*Node)
	return n.items.String()
}

func (f Formatter) FormatEmbeddings(node lattice.Node) []string {
	n := node.(*Node)
	txs := make([]string, 0, len(n.txs))
	for _, tx := range n.txs {
		txs = append(txs, fmt.Sprintf("%v", tx))
	}
	return txs
}
