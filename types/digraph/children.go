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
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/types/digraph/subgraph"
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

func precheckChildren(n Node, kidCount bytes_int.MultiMap, kids bytes_bytes.MultiMap) (has bool, nodes []lattice.Node, err error) {
	dt := n.dt()
	if n.isRoot() {
		nodes, err = n.loadFrequentVertices()
		if err != nil {
			return false, nil, err
		}
		return true, nodes, nil
	}
	if n.edges() >= dt.MaxEdges {
		return true, []lattice.Node{}, nil
	}
	if nodes, has, err := cachedAdj(n, dt, kidCount, kids); err != nil {
		return false, nil, err
	} else if has {
		return true, nodes, nil
	}
	return false, nil, nil
}

func canonChildren(n Node) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if has, nodes, err := precheckChildren(n, dt.CanonKidCount, dt.CanonKids); err != nil {
		return nil, err
	} else if has {
		return nodes, nil
	}
	sg := n.SubGraph()
	nodes, err = findChildren(n, func(pattern *subgraph.SubGraph) (bool, error) {
		return isCanonicalExtension(sg, pattern)
	}, false)
	if err != nil {
		return nil, err
	}
	return nodes, cacheAdj(dt, dt.CanonKidCount, dt.CanonKids, n.Label(), nodes)
}

func children(n Node) (nodes []lattice.Node, err error) {
	dt := n.dt()
	if has, nodes, err := precheckChildren(n, dt.ChildCount, dt.Children); err != nil {
		return nil, err
	} else if has {
		// errors.Logf("DEBUG", "got from precheck %v", n)
		return nodes, nil
	}
	nodes, err = findChildren(n, nil, false)
	if err != nil {
		return nil, err
	}
	return nodes, cacheAdj(dt, dt.ChildCount, dt.Children, n.Label(), nodes)
}

func findChildren(n Node, allow func(*subgraph.SubGraph) (bool, error), debug bool) (nodes []lattice.Node, err error) {
	if debug {
		errors.Logf("CHILDREN-DEBUG", "node %v", n)
	}
	dt := n.dt()
	sg := n.SubGraph()
	patterns, err := extendNode(dt, n, debug)
	if err != nil {
		return nil, err
	}
	unsupExts, err := n.UnsupportedExts()
	if err != nil {
		return nil, err
	}
	newUnsupportedExts := unsupExts.Copy()
	nOverlap, err := n.Overlap()
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	type nodeEp struct {
		n lattice.Node
		vord []int
	}
	nodeCh := make(chan nodeEp)
	vords := make([][]int, 0, 10)
	go func() {
		for nep := range nodeCh {
			nodes = append(nodes, nep.n)
			vords = append(vords, nep.vord)
			wg.Done()
		}
	}()
	epCh := make(chan *subgraph.Extension)
	go func() {
		for ep := range epCh {
			newUnsupportedExts.Add(ep)
			wg.Done()
		}
	}()
	errorCh := make(chan error)
	errs := make([]error, 0, 10)
	go func() {
		for err := range errorCh {
			errs = append(errs, err)
			wg.Done()
		}
	}()
	for k, v, next := patterns.Iterate()(); next != nil; k, v, next = next() {
		err := dt.pool.Do(func(pattern *subgraph.SubGraph, i *extInfo) func() {
			wg.Add(1)
			return func() {
				if allow != nil {
					if allowed, err := allow(pattern); err != nil {
						errorCh <- err
						return
					} else if !allowed {
						wg.Done()
						return
					}
				}
				ep := i.ep
				vord := i.vord
				tu := set.NewSetMap(hashtable.NewLinearHash())
				for i, next := unsupExts.Items()(); next != nil; i, next = next() {
					tu.Add(i.(*subgraph.Extension).Translate(len(sg.V), vord))
				}
				pOverlap := translateOverlap(nOverlap, vord)
				support, exts, embs, overlap, err := ExtsAndEmbs(dt, pattern, pOverlap, tu, dt.Mode, debug)
				if err != nil {
					errorCh <- err
					return
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
		}(k.(*subgraph.SubGraph), v.(*extInfo)))
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()
	close(nodeCh)
	close(epCh)
	close(errorCh)
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
	ep   *subgraph.Extension
	vord []int
}

func extendNode(dt *Digraph, n Node, debug bool) (*hashtable.LinearHash, error) {
	if debug {
		errors.Logf("DEBUG", "n.SubGraph %v", n.SubGraph())
	}
	sg := n.SubGraph()
	b := subgraph.Build(len(sg.V), len(sg.E)).From(sg)
	extPoints, err := n.Extensions()
	if err != nil {
		return nil, err
	}
	patterns := hashtable.NewLinearHash()
	for _, ep := range extPoints {
		bc := b.Copy()
		bc.Extend(ep)
		if len(bc.V) > dt.MaxVertices {
			continue
		}
		vord, eord := bc.CanonicalPermutation()
		ext := bc.BuildFromPermutation(vord, eord)
		if !patterns.Has(ext) {
			patterns.Put(ext, &extInfo{ep, vord})
		}
	}

	return patterns, nil
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
