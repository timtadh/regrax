package digraph

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

type baseLoader struct {
	dt *Digraph
	g *goiso.Graph
	vidxs map[int32]int32
	excluded map[int32]bool
}

func newBaseLoader(dt *Digraph, g *goiso.Graph) *baseLoader {
	return &baseLoader{
		dt: dt,
		g: g,
		vidxs: make(map[int32]int32),
		excluded: make(map[int32]bool),
	}
}

func (l *baseLoader) addVertex(id int32, label string, attrs map[string]interface{}) (err error) {
	if l.dt.Include != nil && !l.dt.Include.MatchString(label) {
		l.excluded[id] = true
		return nil
	}
	if l.dt.Exclude != nil && l.dt.Exclude.MatchString(label) {
		l.excluded[id] = true
		return nil
	}
	vertex := l.g.AddVertex(int(id), label)
	l.vidxs[id] = int32(vertex.Idx)
	err = l.dt.NodeAttrs.Add(int32(vertex.Id), attrs)
	if err != nil {
		return err
	}
	return nil
}

func (l *baseLoader) addEdge(sid, tid int32, label string) (err error) {
	if l.excluded[sid] || l.excluded[tid] {
		return nil
	}
	if l.dt.Include != nil && !l.dt.Include.MatchString(label) {
		return nil
	}
	if l.dt.Exclude != nil && l.dt.Exclude.MatchString(label) {
		return nil
	}
	if sidx, has := l.vidxs[sid]; !has {
		return errors.Errorf("unknown src id %v", tid)
	} else if tidx, has := l.vidxs[tid]; !has{
		return errors.Errorf("unknown targ id %v", tid)
	} else {
		l.g.AddEdge(&l.g.V[sidx], &l.g.V[tidx], label)
	}
	return nil
}

