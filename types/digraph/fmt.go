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
	g     *Digraph
	prfmt lattice.PrFormatter
}

func NewFormatter(g *Digraph, prfmt lattice.PrFormatter) *Formatter {
	return &Formatter{
		g:     g,
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
		if len(n.embeddings) > 0 {
			return n.embeddings[0].Label()
		} else {
			return "0:0"
		}
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
		if len(n.embeddings) > 0 {
			Pat := n.embeddings[0].Label()
			dot := n.embeddings[0].String()
			return fmt.Sprintf("// %s%s\n\n%s\n", Pat, max, dot), nil
		} else {
			return fmt.Sprintf("// 0:0\n\ndigraph{}\n"), nil
		}
	default:
		return "", errors.Errorf("unknown node type %v", node)
	}
}

func (f *Formatter) Embeddings(node lattice.Node) ([]string, error) {
	var embeddings []*goiso.SubGraph = nil
	switch n := node.(type) {
	case *EmbListNode:
		embeddings = n.embeddings
	default:
		return nil, errors.Errorf("unknown node type %v", node)
	}
	embs := make([]string, 0, len(embeddings))
	for _, sg := range embeddings {
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
