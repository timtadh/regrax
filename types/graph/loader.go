package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/stores/bytes_subgraph"
	"github.com/timtadh/sfp/stores/int_json"
)


type ErrorList []error

func (self ErrorList) Error() string {
	var s []string
	for _, err := range self {
		s = append(s, err.Error())
	}
	return "Errors [" + strings.Join(s, ", ") + "]"
}

type MakeLoader func(*Graph) lattice.Loader

type Graph struct {
	G *goiso.Graph
	NodeAttrs int_json.MultiMap
	Embeddings bytes_subgraph.MultiMap
	Parents bytes_subgraph.MultiMap
	ParentCount bytes_int.MultiMap
	Children bytes_subgraph.MultiMap
	ChildCount bytes_int.MultiMap
	FrequentVertices []lattice.Node
	makeLoader MakeLoader
	config *config.Config
}

func NewGraph(config *config.Config, makeLoader MakeLoader) (g *Graph, err error) {
	nodeAttrs, err := config.IntJsonMultiMap("graph-node-attrs")
	if err != nil {
		return nil, err
	}
	childCount, err := config.BytesIntMultiMap("graph-child-count")
	if err != nil {
		return nil, err
	}
	parentCount, err := config.BytesIntMultiMap("graph-parent-count")
	if err != nil {
		return nil, err
	}
	g = &Graph{
		NodeAttrs: nodeAttrs,
		ParentCount: parentCount,
		ChildCount: childCount,
		makeLoader: makeLoader,
		config: config,
	}
	return g, nil
}

func (g *Graph) Loader() lattice.Loader {
	return g.makeLoader(g)
}

func (g *Graph) Close() error {
	return nil
}

type VegLoader struct {
	g *Graph
}

func NewVegLoader(g *Graph) lattice.Loader {
	return &VegLoader{
		g: g,
	}
}

func (v *VegLoader) StartingPoints(input lattice.Input, support int) (nodes []lattice.Node, err error) {
	G, err := v.loadGraph(input)
	if err != nil {
		return nil, err
	}
	v.g.G = G
	v.g.Embeddings, err = v.g.config.BytesSubgraphMultiMap("graph-embeddings", bytes_subgraph.DeserializeSubGraph(G))
	if err != nil {
		return nil, err
	}
	v.g.Parents, err = v.g.config.BytesSubgraphMultiMap("graph-parents", bytes_subgraph.DeserializeSubGraph(G))
	if err != nil {
		return nil, err
	}
	v.g.Children, err = v.g.config.BytesSubgraphMultiMap("graph-children", bytes_subgraph.DeserializeSubGraph(G))
	if err != nil {
		return nil, err
	}

	for i := range G.V {
		u := &G.V[i]
		if G.ColorFrequency(u.Color) >= v.g.config.Support {
			sg := G.SubGraph([]int{u.Idx}, nil)
			err := v.g.Embeddings.Add(sg.ShortLabel(), sg)
			if err != nil {
				return nil, err
			}
		}
	}

	err = bytes_subgraph.DoKey(v.g.Embeddings.Keys, func(label []byte) error {
		sgs := make([]*goiso.SubGraph, 0, 10)
		err := v.g.Embeddings.DoFind(label, func(_ []byte, sg *goiso.SubGraph) error {
			sgs = append(sgs, sg)
			return nil
		})
		if err != nil {
			return err
		}
		if len(sgs) >= support {
			nodes = append(nodes, &Node{label: label, sgs: sgs})
			errors.Logf("INFO", "start %v %v", sgs[0].Label(), len(sgs))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	v.g.FrequentVertices = nodes

	return nodes, nil
}

func (v *VegLoader) loadGraph(input lattice.Input) (graph *goiso.Graph, err error) {
	var errs ErrorList
	V, E, err := graphSize(input)
	if err != nil {
		return nil, err
	}
	G := goiso.NewGraph(V, E)
	graph = &G
	vids := hashtable.NewLinearHash() // int64 ==> *goiso.Vertex

	in, closer := input()
	defer closer()
	err = processLines(in, func(line []byte) {
		if len(line) == 0 || !bytes.Contains(line, []byte("\t")) {
			return
		}
		line_type, data := parseLine(line)
		switch line_type {
		case "vertex":
			if err := v.loadVertex(graph, vids, data); err != nil {
				errs = append(errs, err)
			}
		case "edge":
			if err := v.loadEdge(graph, vids, data); err != nil {
				errs = append(errs, err)
			}
		default:
			errs = append(errs, errors.Errorf("Unknown line type %v", line_type))
		}
	})
	if err != nil {
		return nil, err
	}
	if len(errs) == 0 {
		return graph, nil
	}
	return graph, errs
}

func (v *VegLoader) loadVertex(g *goiso.Graph, vids types.Map, data []byte) (err error) {
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
	if v.g.NodeAttrs != nil {
		err = v.g.NodeAttrs.Add(int32(vertex.Idx), obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VegLoader) loadEdge(g *goiso.Graph, vids types.Map, data []byte) (err error) {
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

func processLines(in io.Reader, process func([]byte)) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		unsafe := scanner.Bytes()
		line := make([]byte, len(unsafe))
		copy(line, unsafe)
		process(line)
	}
	return scanner.Err()
}

func parseJson(data []byte) (obj map[string]interface{}, err error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func parseLine(line []byte) (line_type string, data []byte) {
	split := bytes.Split(line, []byte("\t"))
	return strings.TrimSpace(string(split[0])), bytes.TrimSpace(split[1])
}

func graphSize(input lattice.Input) (V, E int, err error) {
	in, closer := input()
	defer closer()
	err = processLines(in, func(line []byte) {
		if bytes.HasPrefix(line, []byte("vertex")) {
			V++
		} else if bytes.HasPrefix(line, []byte("edge")) {
			E++
		}
	})
	if err != nil {
		return 0, 0, err
	}
	return V, E, nil
}

