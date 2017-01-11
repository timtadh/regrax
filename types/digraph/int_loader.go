package digraph

import (
	"bytes"
	"strings"
	"strconv"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/digraph"
)

type IntLoader struct {
	dt *Digraph
}

func NewIntLoader(config *config.Config, dc *Config) (lattice.Loader, error) {
	g, err := NewDigraph(config, dc)
	if err != nil {
		return nil, err
	}
	v := &IntLoader{
		dt: g,
	}
	return v, nil
}

func (v *IntLoader) Load(input lattice.Input) (dt lattice.DataType, err error) {
	return v.LoadWithLabels(input, digraph.NewLabels())
}

func (v *IntLoader) LoadWithLabels(input lattice.Input, labels *digraph.Labels) (lattice.DataType, error) {
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

func (v *IntLoader) loadDigraph(input lattice.Input, labels *digraph.Labels) (*digraph.Builder, error) {
	var errs ErrorList
	V, E, err := intGraphSize(input)
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "Got graph size %v %v", V, E)
	G := digraph.Build(V, E)
	b := newBaseLoader(v.dt, G)

	in, closer := input()
	defer closer()
	err = processLines(in, func(line []byte) {
		if len(line) == 0 || bytes.Contains(line, []byte("#")) {
			return
		}
		line_type, data := intParseLine(line)
		switch line_type {
		case "v":
			if err := v.loadVertex(labels, b, data); err != nil {
				errs = append(errs, err)
			}
		case "e":
			if err := v.loadEdge(labels, b, data); err != nil {
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
		return G, nil
	}
	return nil, errs
}

func (v *IntLoader) loadVertex(labels *digraph.Labels, b *baseLoader, data []byte) (err error) {
	split := bytes.SplitN(data, []byte(" "), 2)
	id, err := strconv.Atoi(string(split[0]))
	if err != nil {
		return err
	}
	label := string(split[1])
	return b.addVertex(int32(id), labels.Color(label), label, nil)
}

func (v *IntLoader) loadEdge(labels *digraph.Labels, b *baseLoader, data []byte) (err error) {
	split := bytes.SplitN(data, []byte(" "), 3)
	src, err := strconv.Atoi(string(split[0]))
	if err != nil {
		return err
	}
	targ, err := strconv.Atoi(string(split[1]))
	if err != nil {
		return err
	}
	label := string(split[2])
	return b.addEdge(int32(src), int32(targ), labels.Color(label), label )
}

func intParseLine(line []byte) (line_type string, data []byte) {
	split := bytes.SplitN(line, []byte(" "), 2)
	return strings.TrimSpace(string(split[0])), bytes.TrimSpace(split[1])
}

func intGraphSize(input lattice.Input) (V, E int, err error) {
	in, closer := input()
	defer closer()
	err = processLines(in, func(line []byte) {
		if bytes.HasPrefix(line, []byte("v")) {
			V++
		} else if bytes.HasPrefix(line, []byte("e")) {
			E++
		}
	})
	if err != nil {
		return 0, 0, err
	}
	return V, E, nil
}
