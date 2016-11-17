package digraph

import (
	"fmt"
	"io"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/subgraph"
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
			return n.Pat.Pretty(n.Dt.Labels)
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
		if len(n.embeddings) > 0 {
			Pat := n.Pat.Pretty(n.Dt.Labels)
			allAttrs, err := f.loadAttrs(n.embeddings[0])
			if err != nil {
				return "", err
			}
			attrs := make(map[int]map[string]interface{})
			for id, _ := range allAttrs {
				attrs[id] = make(map[string]interface{})
				if size, has := allAttrs[id]["fontsize"]; has {
					attrs[id]["fontsize"] = size
				}
			}
			dot := n.embeddings[0].Dotty(n.Dt.Labels, attrs)
			return fmt.Sprintf("// %s\n\n%s\n", Pat, dot), nil
		} else {
			return fmt.Sprintf("// {0:0}\n\ndigraph{}\n"), nil
		}
	default:
		return "", errors.Errorf("unknown node type %v", node)
	}
}

func (f *Formatter) Embeddings(node lattice.Node) ([]string, error) {
	var dt *Digraph
	var embeddings []*subgraph.Embedding = nil
	switch n := node.(type) {
	case *EmbListNode:
		embeddings = n.embeddings
		dt = n.Dt
	default:
		return nil, errors.Errorf("unknown node type %v", node)
	}
	embs := make([]string, 0, len(embeddings))
	for _, emb := range embeddings {
		allAttrs, err := f.loadAttrs(emb)
		if err != nil {
			return nil, err
		}
		embs = append(embs, emb.Dotty(dt.Labels, allAttrs))
	}
	return embs, nil
}

func (f *Formatter) loadAttrs(emb *subgraph.Embedding) (map[int]map[string]interface{}, error) {
	allAttrs := make(map[int]map[string]interface{})
	for _, id := range emb.Ids {
		err := f.g.NodeAttrs.DoFind(
			int32(id),
			func(_ int32, attrs map[string]interface{}) error {
				allAttrs[id] = attrs
				return nil
			})
		if err != nil {
			return nil, err
		}
	}
	return allAttrs, nil
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
