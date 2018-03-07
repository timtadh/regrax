package digraph

import (
	"fmt"
	"io/ioutil"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/dot"
	"github.com/timtadh/combos"
)

import (
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/types/digraph/digraph"
)

type DotLoader struct {
	dt *Digraph
}

func NewDotLoader(config *config.Config, dc *Config) (lattice.Loader, error) {
	g, err := NewDigraph(config, dc)
	if err != nil {
		return nil, err
	}
	v := &DotLoader{
		dt: g,
	}
	return v, nil
}

func (v *DotLoader) Load(input lattice.Input) (dt lattice.DataType, err error) {
	return v.LoadWithLabels(input, digraph.NewLabels())
}

func (v *DotLoader) LoadWithLabels(input lattice.Input, labels *digraph.Labels) (lattice.DataType, error) {
	G, err := v.loadDigraph(input, labels)
	if err != nil {
		return nil, err
	}
	err = v.dt.Init(G, labels)
	if err != nil {
		return nil, err
	}
	return v.dt, nil
}

func (v *DotLoader) loadDigraph(input lattice.Input, labels *digraph.Labels) (graph *digraph.Builder, err error) {
	r, closer := input()
	text, err := ioutil.ReadAll(r)
	closer()
	if err != nil {
		return nil, err
	}
	G := digraph.Build(100, 1000)
	dp := &dotParse{
		b: newBaseLoader(v.dt, G),
		d: v,
		labels: labels,
		vids: make(map[string]int32),
	}
	// s, err := dot.Lexer.Scanner(text)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, _, eof := s.Next(); !eof; _, _, eof = s.Next() {}
	err = dot.StreamParse(text, dp)
	if err != nil {
		return nil, err
	}
	return G, nil
}

type dotParse struct {
	b *baseLoader
	d *DotLoader
	labels *digraph.Labels
	graphId int
	curGraph string
	subgraph int
	nextId int32
	vids map[string]int32
}

func (p *dotParse) Enter(name string, n *combos.Node) error {
	if name == "SubGraph" {
		p.subgraph += 1
		return nil
	}
	p.curGraph = fmt.Sprintf("%v-%d", n.Get(1).Value.(string), p.graphId)
	// errors.Logf("DEBUG", "enter %v %v", p.curGraph, n)
	return nil
}

func (p *dotParse) Stmt(n *combos.Node) error {
	if false {
		errors.Logf("DEBUG", "stmt %v", n)
	}
	if p.subgraph > 0 {
		return nil
	}
	switch n.Label {
	case "Node":
		p.loadVertex(n)
		// errors.Logf("DEBUG", "node %v", n)
	case "Edge":
		p.loadEdge(n)
		// errors.Logf("DEBUG", "edge %v", n)
	}
	return nil
}

func (p *dotParse) Exit(name string) error {
	if name == "SubGraph" {
		p.subgraph--
		return nil
	}
	p.graphId++
	return nil
}

func (p *dotParse) loadVertex(n *combos.Node) (err error) {
	sid := n.Get(0).Value.(string)
	attrs := make(map[string]interface{})
	for _, attr := range n.Get(1).Children {
		name := attr.Get(0).Value.(string)
		value := attr.Get(1).Value.(string)
		attrs[name] = value
	}
	attrs["graphId"] = p.graphId
	id := p.nextId
	p.nextId++
	p.vids[sid] = id
	label := sid
	if l, has := attrs["label"]; has {
		label = l.(string)
	}
	return p.b.addVertex(id, p.labels.Color(label), label, attrs)
}

func (p *dotParse) loadEdge(n *combos.Node) (err error) {
	getId := func(sid string) (int32, error) {
		if _, has := p.vids[sid]; !has {
			err := p.loadVertex(combos.NewNode("Node").
				AddKid(combos.NewValueNode("ID", sid)).
				AddKid(combos.NewNode("Attrs")))
			if err != nil {
				return 0, err
			}
		}
		return p.vids[sid], nil
	}
	srcSid := n.Get(0).Value.(string)
	sid, err := getId(srcSid)
	if err != nil {
		return err
	}
	targSid := n.Get(1).Value.(string)
	tid, err := getId(targSid)
	if err != nil {
		return err
	}
	label := ""
	for _, attr := range n.Get(2).Children {
		name := attr.Get(0).Value.(string)
		if name == "label" {
			label = attr.Get(1).Value.(string)
			break
		}
	}
	return p.b.addEdge(sid, tid, p.labels.Color(label), label)
}
