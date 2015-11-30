package graph

import (
	"log"
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

func (f *Formatter) FormatPattern(node lattice.Node) string {
	n := node.(*Node)
	return n.sgs[0].String()
}

func (f *Formatter) FormatEmbeddings(node lattice.Node) []string {
	n := node.(*Node)
	embs := make([]string, 0, len(n.sgs))
	for _, sg := range n.sgs {
		allAttrs := make(map[int]map[string]interface{})
		for _, v := range sg.V {
			err := f.g.NodeAttrs.DoFind(
				int32(v.Id),
				func(_ int32, attrs map[string]interface{}) error {
					allAttrs[v.Id] = attrs
					return nil
				})
			if err != nil {
				log.Fatal(err)
			}
		}
		embs = append(embs, sg.StringWithAttrs(allAttrs))
	}
	return embs
}


