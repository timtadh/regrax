package digraph

import (
	"fmt"
	"io"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Formatter struct {
	g *Graph
	prfmt lattice.PrFormatter
}

func NewFormatter(g *Graph, prfmt lattice.PrFormatter) *Formatter {
	return &Formatter{
		g: g,
		prfmt: prfmt,
	}
}

func (f *Formatter) PrFormatter() lattice.PrFormatter {
	return f.prfmt
}

func (f *Formatter) FileExt() string {
	return ".dot"
}

func (f *Formatter) PatternName(node lattice.Node) string {
	switch n := node.(type) {
	case *EmbListNode:
		if len(n.sgs) > 0 {
			return n.sgs[0].Label()
		} else {
			return "0:0"
		}
	case *SearchNode:
		return n.pat.String()
	default:
		panic(errors.Errorf("unknown node type %v", node))
	}
}

func (f *Formatter) Pattern(node lattice.Node) (string, error) {
	switch n := node.(type) {
	case *EmbListNode:
		max := ""
		if ismax, err := n.Maximal(); err != nil {
			return "", err
		} else if ismax {
			max = " # maximal"
		}
		if len(n.sgs) > 0 {
			pat := n.sgs[0].Label()
			dot := n.sgs[0].String()
			return fmt.Sprintf("// %s%s\n\n%s\n", pat, max, dot), nil
		} else {
			return fmt.Sprintf("// 0:0\n\ndigraph{}\n", ), nil
		}
	case *SearchNode:
		max := ""
		if ismax, err := n.Maximal(); err != nil {
			return "", err
		} else if ismax {
			max = " # maximal"
		}
		pat := n.pat.String()
		embs, err := n.Embeddings()
		if err != nil {
			return "", err
		}
		dot := embs[0].String()
		return fmt.Sprintf("// %s%s\n\n%s\n", pat, max, dot), nil
	default:
		return "", errors.Errorf("unknown node type %v", node)
	}
}

func (f *Formatter) Embeddings(node lattice.Node) ([]string, error) {
	var sgs []*goiso.SubGraph = nil
	switch n := node.(type) {
	case *EmbListNode:
		sgs = n.sgs
	case *SearchNode:
		embs, err := n.Embeddings()
		if err != nil {
			return nil, err
		}
		sgs = embs
	default:
		return nil, errors.Errorf("unknown node type %v", node)
	}
	embs := make([]string, 0, len(sgs))
	for _, sg := range sgs {
		allAttrs := make(map[int]map[string]interface{})
		for _, v := range sg.V {
			err := f.g.NodeAttrs.DoFind(
				int32(f.g.G.V[v.Id].Id),
				func(id int32, attrs map[string]interface{}) error {
					allAttrs[v.Id] = attrs
					return nil
				})
			if err != nil {
				return nil, err
			}
		}
		embs = append(embs, sg.StringWithAttrs(allAttrs))
	}
	return embs, nil
}

func (f *Formatter) FormatPattern(w io.Writer, node lattice.Node) error {
	pat, err := f.Pattern(node)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", pat)
	return err
}

func (f *Formatter) FormatEmbeddings(w io.Writer, node lattice.Node) error {
	embs, err := f.Embeddings(node)
	if err != nil {
		return err
	}
	pat := f.PatternName(node)
	embeddings := strings.Join(embs, "\n")
	_, err = fmt.Fprintf(w, "// %s\n\n%s\n\n", pat, embeddings)
	return err
}

