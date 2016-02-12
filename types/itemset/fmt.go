package itemset

import (
	"fmt"
	"io"
	"strings"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Formatter struct{
	PrFmt lattice.PrFormatter
}

func (f *Formatter) PrFormatter() lattice.PrFormatter {
	return f.PrFmt
}

func (f *Formatter) FileExt() string {
	return ".items"
}

func (f *Formatter) PatternName(node lattice.Node) string {
	n := node.(*Node)
	items := make([]string, 0, n.pat.Items.Size())
	for i, next := n.pat.Items.Items()(); next != nil; i, next = next() {
		items = append(items, fmt.Sprintf("%v", i))
	}
	return fmt.Sprintf("%s", strings.Join(items, " "))
}

func (f *Formatter) Pattern(node lattice.Node) (string, error) {
	return f.PatternName(node), nil
}

func (f *Formatter) Embeddings(node lattice.Node) ([]string, error) {
	n := node.(*Node)
	txs := make([]string, 0, len(n.txs))
	for _, tx := range n.txs {
		txs = append(txs, fmt.Sprintf("%v", tx))
	}
	return txs, nil
}

func (f *Formatter) FormatPattern(w io.Writer, node lattice.Node) error {
	n := node.(*Node)
	pat, err := f.Pattern(node)
	if err != nil {
		return err
	}
	max := ""
	if ismax, err := n.Maximal(); err != nil {
		return err
	} else if ismax {
		max = " # maximal"
	}
	_, err = fmt.Fprintf(w, "%s%s\n", pat, max)
	return err
}

func (f *Formatter) FormatEmbeddings(w io.Writer, node lattice.Node) error {
	txs, err := f.Embeddings(node)
	if err != nil {
		return err
	}
	pat, err := f.Pattern(node)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s : %s\n", pat, strings.Join(txs, " "))
	return err
}
