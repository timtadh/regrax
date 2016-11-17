package uniprox

import (
	"log"
	"math"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners/fastmax"
	"github.com/timtadh/sfp/reporters"
	"github.com/timtadh/sfp/types/digraph"
	dg "github.com/timtadh/sfp/types/digraph/digraph"
	"github.com/timtadh/sfp/types/itemset"
)

func CommonAncestor(patterns []lattice.Pattern) (_ lattice.Pattern, err error) {
	if len(patterns) == 0 {
		return nil, errors.Errorf("no patterns given")
	} else if len(patterns) == 1 {
		return patterns[0], nil
	}
	switch patterns[0].(type) {
	case *digraph.SubgraphPattern:
		return digraphCommonAncestor(patterns)
	case *itemset.Pattern:
		return itemsetCommonAncestor(patterns)
	default:
		return nil, errors.Errorf("unknown pattern type %v", patterns[0])
	}
}

func itemsetCommonAncestor(patterns []lattice.Pattern) (_ lattice.Pattern, err error) {
	var items types.Set
	for i, pat := range patterns {
		p := pat.(*itemset.Pattern)
		if i == 0 {
			items = p.Items
		} else {
			items, err = items.Intersect(p.Items)
			if err != nil {
				return nil, err
			}
		}
	}
	return &itemset.Pattern{items.(*set.SortedSet)}, nil
}

func digraphCommonAncestor(patterns []lattice.Pattern) (lattice.Pattern, error) {

	// construct a in memory configuration for finding common subdigraphs of all patterns
	conf := &config.Config{
		Cache:   "",
		Output:  "",
		Support: len(patterns),
		Samples: 5,
		Unique:  false,
	}
	wlkr := fastmax.NewWalker(conf)
	wlkr.Reject = false

	// closing the walker releases the memory
	defer func() {
		err := wlkr.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	maxE := int(math.MaxInt32)
	maxV := int(math.MaxInt32)
	for _, pat := range patterns {
		sg := pat.(*digraph.SubgraphPattern).Pat
		if len(sg.E) < maxE {
			maxE = len(sg.E)
		}
		if len(sg.V) < maxV {
			maxV = len(sg.V)
		}
	}

	// init the datatype (we are now ready to mine)
	dt, err := digraph.NewDigraph(conf, &digraph.Config{
		MinEdges: 0,
		MaxEdges: maxE,
		MinVertices: 0,
		MaxVertices: maxV,
		Mode: digraph.GIS | digraph.OverlapPruning | digraph.ExtFromEmb,
	})
	if err != nil {
		return nil, err
	}

	var labels *dg.Labels = patterns[0].(*digraph.SubgraphPattern).Dt.Labels
	// construct the digraph from the patterns
	b := dg.Build(10, 10)
	offset := 0
	for gid, pat := range patterns {
		sn := pat.(*digraph.SubgraphPattern)
		for i := range sn.Pat.V {
			vid := offset + i
			b.AddVertex(sn.Pat.V[i].Color)
			err := dt.NodeAttrs.Add(int32(vid), map[string]interface{}{"gid": gid})
			if err != nil {
				return nil, err
			}
		}
		for i := range sn.Pat.E {
			b.AddEdge(&b.V[offset+sn.Pat.E[i].Src], &b.V[offset+sn.Pat.E[i].Targ], sn.Pat.E[i].Color)
		}
		offset += len(sn.Pat.V)
	}

	// Initialize the *Digraph with the graph G being used.
	err = dt.Init(b, labels)
	if err != nil {
		return nil, err
	}

	// errors.Logf("DEBUG", "patterns %v %v", len(patterns), G)

	// create the reporter
	fmtr := digraph.NewFormatter(dt, nil)
	collector := &reporters.Collector{make([]lattice.Node, 0, 10)}
	uniq, err := reporters.NewUnique(conf, fmtr, collector, "")
	if err != nil {
		return nil, err
	}
	// rptr := &reporters.Chain{[]miners.Reporter{reporters.NewLog(fmtr, false, "DEBUG", "common-ancestor"), uniq}}
	rptr := uniq

	// mine
	err = wlkr.Mine(dt, rptr, fmtr)
	if err != nil {
		return nil, err
	}

	// extract the largest common subdigraph
	maxLevel := collector.Nodes[0].Pattern().Level()
	maxPattern := collector.Nodes[0].Pattern()
	for _, n := range collector.Nodes {
		p := n.Pattern()
		if p.Level() > maxLevel {
			maxLevel = p.Level()
			maxPattern = p
		}
	}
	errors.Logf("DEBUG", "ancestor %v", maxPattern)

	return maxPattern, nil
}
