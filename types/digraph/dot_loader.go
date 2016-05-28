package digraph

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
	"github.com/timtadh/dot"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type DotLoader struct {
	dt *Digraph
}

func NewDotLoader(config *config.Config, mode Mode, minE, maxE, minV, maxV int) (lattice.Loader, error) {
	g, err := NewDigraph(config, mode, minE, maxE, minV, maxV)
	if err != nil {
		return nil, err
	}
	v := &DotLoader{
		dt: g,
	}
	return v, nil
}

func (v *DotLoader) Load(input lattice.Input) (dt lattice.DataType, err error) {
	G, err := v.loadDigraph(input)
	if err != nil {
		return nil, err
	}
	err = v.dt.Init(G)
	if err != nil {
		return nil, err
	}
	return v.dt, nil
}

func (v *DotLoader) loadDigraph(input lattice.Input) (graph *goiso.Graph, err error) {
	r, closer := input()
	text, err := ioutil.ReadAll(r)
	closer()
	if err != nil {
		return nil, err
	}
	G := goiso.NewGraph(10, 10)
	dp := &dotParse{
		g: &G,
		d: v,
	}
	err = dot.StreamParse(text, dp)
	if err != nil {
		return nil, err
	}
	return dp.g, nil
}

type dotParse struct {
	g *goiso.Graph
	d *DotLoader
	curGraph string
	subgraph int
}

func (p *dotParse) Enter(name string, n *dot.Node) error {
	if name == "SubGraph" {
		p.subgraph += 1
		return nil
	}
	p.curGraph = n.Get(1).Value.(string)
	return nil
}

func (p *dotParse) Stmt(n *dot.Node) error {
	if p.subgraph > 0 {
		return nil
	}
	if false {
		errors.Logf("DEBUG", "stmt %v", n)
	}
	switch n.Label {
	case "Node":
		// errors.Logf("DEBUG", "node %v", n)
	case "Edge":
		// errors.Logf("DEBUG", "edge %v", n)
	}
	return nil
}

func (p *dotParse) Exit(name string) error {
	if name == "SubGraph" {
		p.subgraph -= 1
		return nil
	}
	return nil
}

func (v *DotLoader) loadVertex(g *goiso.Graph, vids types.Map, data []byte) (err error) {
	obj, err := parseJson(data)
	if err != nil {
		return err
	}
	_id, err := obj["id"].(json.Number).Int64()
	if err != nil {
		return err
	}
	label := strings.TrimSpace(obj["label"].(string))
	id := int(_id)
	vertex := g.AddVertex(id, label)
	err = vids.Put(types.Int(id), vertex)
	if err != nil {
		return err
	}
	err = v.dt.NodeAttrs.Add(int32(vertex.Id), obj)
	if err != nil {
		return err
	}
	return nil
}

func (v *DotLoader) loadEdge(g *goiso.Graph, vids types.Map, data []byte) (err error) {
	obj, err := parseJson(data)
	if err != nil {
		return err
	}
	_src, err := obj["src"].(json.Number).Int64()
	if err != nil {
		return err
	}
	_targ, err := obj["targ"].(json.Number).Int64()
	if err != nil {
		return err
	}
	src := int(_src)
	targ := int(_targ)
	label := strings.TrimSpace(obj["label"].(string))
	if o, err := vids.Get(types.Int(src)); err != nil {
		return err
	} else {
		u := o.(*goiso.Vertex)
		if o, err := vids.Get(types.Int(targ)); err != nil {
			return err
		} else {
			v := o.(*goiso.Vertex)
			g.AddEdge(u, v, label)
		}
	}
	return nil
}
