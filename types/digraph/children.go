package digraph

import (
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/hashtable"
	"github.com/timtadh/data-structures/set"
)

import (
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/types/digraph/subgraph"
)

type Node interface {
	lattice.Node
	New(*subgraph.SubGraph, []*subgraph.Extension, []*subgraph.Embedding, []map[int]bool) Node
	Label() []byte
	Extensions() ([]*subgraph.Extension, error)
	Embeddings() ([]*subgraph.Embedding, error)
	Overlap() ([]map[int]bool, error)
	UnsupportedExts() (*set.SortedSet, error)
	SaveUnsupportedExts(int, []int, *set.SortedSet) error
	SubGraph() *subgraph.SubGraph
	loadFrequentVertices() ([]lattice.Node, error)
	isRoot() bool
	edges() int
	dt() *Digraph
}

func precheckChildren(n Node) (has bool, nodes []lattice.Node, err error) {
	if n.isRoot() {
		nodes, err = n.loadFrequentVertices()
		if err != nil {
			return false, nil, err
		}
		return true, nodes, nil
	}
	return false, nil, nil
}

func canonChildren(n Node) (nodes []lattice.Node, err error) {
	if n.isRoot() {
		return n.loadFrequentVertices()
	}
	sg := n.SubGraph()
	nodes, err = findChildren(n, func(pattern *subgraph.SubGraph) (bool, error) {
		return isCanonicalExtension(sg, pattern)
	}, false)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func children(n Node) (nodes []lattice.Node, err error) {
	if n.isRoot() {
		return n.loadFrequentVertices()
	}
	nodes, err = findChildren(n, nil, false)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func findChildren(n Node, allow func(*subgraph.SubGraph) (bool, error), debug bool) (nodes []lattice.Node, err error) {
	if debug {
		errors.Logf("CHILDREN-DEBUG", "node %v", n)
	}
	dt := n.dt()
	sg := n.SubGraph()
	extPoints, err := n.Extensions()
	if err != nil {
		return nil, err
	}
	patterns := make(chan *extInfo, 100)
	go extendNode(dt, n, extPoints, patterns)
	unsupExts, err := n.UnsupportedExts()
	if err != nil {
		return nil, err
	}
	newUnsupportedExts := unsupExts.Copy()
	nOverlap, err := n.Overlap()
	if err != nil {
		return nil, err
	}
	type nodeEp struct {
		n    lattice.Node
		vord []int
	}
	nodeCh := make(chan nodeEp)
	vords := make([][]int, 0, 10)
	var dataWg sync.WaitGroup
	dataWg.Add(3)
	go func() {
		for nep := range nodeCh {
			nodes = append(nodes, nep.n)
			vords = append(vords, nep.vord)
		}
		dataWg.Done()
	}()
	epCh := make(chan *subgraph.Extension)
	go func() {
		for ep := range epCh {
			newUnsupportedExts.Add(ep)
		}
		dataWg.Done()
	}()
	errorCh := make(chan error)
	errs := make([]error, 0, 10)
	go func() {
		for err := range errorCh {
			errs = append(errs, err)
		}
		dataWg.Done()
	}()
	workers := dt.config.Workers()
	var workersWg sync.WaitGroup
	workersWg.Add(workers)
	for x := 0; x < workers; x++ {
		go func(tid int) {
			for i := range patterns {
				pattern := i.ext
				ep := i.ep
				vord := i.vord
				if allow != nil {
					if allowed, err := allow(pattern); err != nil {
						errorCh <- err
						break
					} else if !allowed {
						continue
					}
				}
				tu := set.NewSetMap(hashtable.NewLinearHash())
				for i, next := unsupExts.Items()(); next != nil; i, next = next() {
					tu.Add(i.(*subgraph.Extension).Translate(len(sg.V), vord))
				}
				pOverlap := translateOverlap(nOverlap, vord)
				support, exts, embs, overlap, err := ExtsAndEmbs(dt, pattern, pOverlap, tu, dt.Mode, debug)
				if err != nil {
					errorCh <- err
					break
				}
				if debug {
					errors.Logf("CHILDREN-DEBUG", "pattern %v support %v exts %v", pattern.Pretty(dt.Labels), len(embs), len(exts))
				}
				if support >= dt.Support() {
					nodeCh <- nodeEp{n.New(pattern, exts, embs, overlap), vord}
				} else {
					epCh <- ep
				}
			}
			workersWg.Done()
		}(x)
	}
	workersWg.Wait()
	close(nodeCh)
	close(epCh)
	close(errorCh)
	dataWg.Wait()
	if len(errs) > 0 {
		e := errors.Errorf("findChildren error").(*errors.Error)
		for _, err := range errs {
			e.Chain(err)
		}
		return nil, e
	}
	for i, newNode := range nodes {
		err := newNode.(Node).SaveUnsupportedExts(len(sg.V), vords[i], newUnsupportedExts)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

type extInfo struct {
	ext  *subgraph.SubGraph
	ep   *subgraph.Extension
	vord []int
}

func extendNode(dt *Digraph, n Node, extPoints []*subgraph.Extension, ch chan *extInfo) {
	sg := n.SubGraph()
	b := subgraph.Build(len(sg.V), len(sg.E)).From(sg)
	patterns := hashtable.NewLinearHash()
	for _, ep := range extPoints {
		bc := b.Copy()
		bc.Extend(ep)
		if len(bc.V) <= dt.MaxVertices {
			vord, eord := bc.CanonicalPermutation()
			ext := bc.BuildFromPermutation(vord, eord)
			if !patterns.Has(ext) {
				patterns.Put(ext, nil)
				ch <- &extInfo{ext, ep, vord}
			}
		}
	}
	close(ch)
}

func translateOverlap(org []map[int]bool, vord []int) []map[int]bool {
	if org == nil {
		return nil
	}
	neo := make([]map[int]bool, len(vord))
	for idx, o := range org {
		neo[vord[idx]] = o
	}
	return neo
}
