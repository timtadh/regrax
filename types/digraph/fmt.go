package digraph

import (
	"fmt"
	"io"
	"strings"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Formatter struct {
	g *Graph
}

func NewFormatter(g *Graph) *Formatter {
	return &Formatter{
		g: g,
	}
}

func (f *Formatter) FileExt() string {
	return ".dot"
}

func (f *Formatter) PatternName(node lattice.Node) string {
	n := node.(*Node)
	return n.sgs[0].Label()
}

func (f *Formatter) Pattern(node lattice.Node) (string, error) {
	n := node.(*Node)
	max := ""
	if ismax, err := n.Maximal(); err != nil {
		return "", err
	} else if ismax {
		max = " # maximal"
	}
	pat := n.sgs[0].Label()
	dot := n.sgs[0].String()
	return fmt.Sprintf("// %s%s\n\n%s\n", pat, max, dot), nil
}

func (f *Formatter) Embeddings(node lattice.Node) (string, error) {
	n := node.(*Node)
	pat := n.sgs[0].Label()
	embs := make([]string, 0, len(n.sgs))
	for _, sg := range n.sgs {
		allAttrs := make(map[int]map[string]interface{})
		for _, v := range sg.V {
			err := f.g.NodeAttrs.DoFind(
				int32(f.g.G.V[v.Id].Id),
				func(id int32, attrs map[string]interface{}) error {
					allAttrs[v.Id] = attrs
					return nil
				})
			if err != nil {
				return "", err
			}
		}
		embs = append(embs, sg.StringWithAttrs(allAttrs))
	}
	embeddings := strings.Join(embs, "\n")
	return fmt.Sprintf("// %s\n\n%s\n", pat, embeddings), nil
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
	emb, err := f.Embeddings(node)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", emb)
	return err
}

