package digraph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/types/digraph/digraph"
)

type ErrorList []error

func (self ErrorList) Error() string {
	var s []string
	for _, err := range self {
		s = append(s, err.Error())
	}
	return "Errors [" + strings.Join(s, ", ") + "]"
}

type VegLoader struct {
	dt *Digraph
}

func NewVegLoader(config *config.Config, dc *Config) (lattice.Loader, error) {
	g, err := NewDigraph(config, dc)
	if err != nil {
		return nil, err
	}
	v := &VegLoader{
		dt: g,
	}
	return v, nil
}

func (v *VegLoader) Load(input lattice.Input) (dt lattice.DataType, err error) {
	return v.LoadWithLabels(input, digraph.NewLabels())
}

func (v *VegLoader) LoadWithLabels(input lattice.Input, labels *digraph.Labels) (lattice.DataType, error) {
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

func (v *VegLoader) loadDigraph(input lattice.Input, labels *digraph.Labels) (*digraph.Builder, error) {
	var errs ErrorList
	V, E, err := graphSize(input)
	if err != nil {
		return nil, err
	}
	errors.Logf("DEBUG", "Got graph size %v %v", V, E)
	G := digraph.Build(V, E)
	b := newBaseLoader(v.dt, G)

	in, closer := input()
	defer closer()
	err = processLines(in, func(line []byte) {
		if len(line) == 0 || !bytes.Contains(line, []byte("\t")) {
			return
		}
		line_type, data := parseLine(line)
		switch line_type {
		case "vertex":
			if err := v.loadVertex(labels, b, data); err != nil {
				errs = append(errs, err)
			}
		case "edge":
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

func (v *VegLoader) loadVertex(labels *digraph.Labels, b *baseLoader, data []byte) (err error) {
	obj, err := parseJson(data)
	if err != nil {
		return err
	}
	_id, err := obj["id"].(json.Number).Int64()
	if err != nil {
		return err
	}
	label := strings.TrimSpace(obj["label"].(string))
	id := int32(_id)
	return b.addVertex(id, labels.Color(label), label, obj)
}

func (v *VegLoader) loadEdge(labels *digraph.Labels, b *baseLoader, data []byte) (err error) {
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
	src := int32(_src)
	targ := int32(_targ)
	label := strings.TrimSpace(obj["label"].(string))
	return b.addEdge(src, targ, labels.Color(label), label)
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
