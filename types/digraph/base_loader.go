package digraph

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/types/digraph/digraph"
)

type baseLoader struct {
	dt *Digraph
	b *digraph.Builder
	vidxs map[int32]int32
	excluded map[int32]bool
}

func newBaseLoader(dt *Digraph, b *digraph.Builder) *baseLoader {
	return &baseLoader{
		dt: dt,
		b: b,
		vidxs: make(map[int32]int32),
		excluded: make(map[int32]bool),
	}
}

func (l *baseLoader) addVertex(id int32, color int, label string, attrs map[string]interface{}) (err error) {
	if l.dt.Include != nil && !l.dt.Include.MatchString(label) {
		l.excluded[id] = true
		return nil
	}
	if l.dt.Exclude != nil && l.dt.Exclude.MatchString(label) {
		l.excluded[id] = true
		return nil
	}
	vertex := l.b.AddVertex(color)
	l.vidxs[id] = int32(vertex.Idx)
	if l.dt.NodeAttrs != nil && attrs != nil {
		attrs["oid"] = id
		attrs["color"] = color
		err = l.dt.NodeAttrs.Add(int32(vertex.Idx), attrs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *baseLoader) addEdge(sid, tid int32, color int, label string) (err error) {
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
		l.b.AddEdge(&l.b.V[sidx], &l.b.V[tidx], color)
	}
	return nil
}

