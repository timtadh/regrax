package uniprox

import (
	"log"
	"math"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/fastmax"
	"github.com/timtadh/sfp/miners/reporters"
	"github.com/timtadh/sfp/types/digraph"
	"github.com/timtadh/sfp/types/itemset"
)

func CommonAncestor(patterns []lattice.Pattern) (_ lattice.Pattern, err error) {
	if len(patterns) == 0 {
		return nil, errors.Errorf("no patterns given")
	} else if len(patterns) == 1 {
		return patterns[0], nil
	}
	switch patterns[0].(type) {
	case *digraph.Pattern: return digraphCommonAncestor(patterns)
	case *itemset.Pattern: return itemsetCommonAncestor(patterns)
	default: return nil, errors.Errorf("unknown pattern type %v", patterns[0])
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
		Support: len(patterns),
		Samples: 5,
		Unique: false,
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
		sg := pat.(*digraph.Pattern).Sg
		if len(sg.E) < maxE {
			maxE = len(sg.E)
		}
		if len(sg.V) < maxV {
			maxV = len(sg.V)
		}
	}

	// construct the digraph from the patterns
	Graph := goiso.NewGraph(10, 10)
	G := &Graph
	l, err := digraph.NewVegLoader(conf, false, digraph.MinImgSupported, 0, maxE, 0, maxV)
	if err != nil {
		return nil, err
	}
	v := l.(*digraph.VegLoader)
	offset := 0
	for _, pat := range patterns {
		sg := pat.(*digraph.Pattern).Sg
		for i := range sg.V {
			G.AddVertex(offset + i, sg.G.Colors[sg.V[i].Color])
		}
		for i := range sg.E {
			G.AddEdge(&G.V[offset + sg.E[i].Src], &G.V[offset + sg.E[i].Targ], sg.G.Colors[sg.E[i].Color])
		}
		offset += len(sg.V)
	}

	// compute the starting points (we are now ready to mine)
	start, err := v.ComputeStartingPoints(G)
	if err != nil {
		return nil, err
	}
	v.G.FrequentVertices = start
	var dt lattice.DataType = v.G

	errors.Logf("DEBUG", "patterns %v %v", len(patterns), G)

	// create the reporter
	fmtr := digraph.NewFormatter(v.G, nil)
	collector := &reporters.Collector{make([]lattice.Node, 0, 10)}
	uniq, err := reporters.NewUnique(conf, fmtr, collector, "")
	if err != nil {
		return nil, err
	}
	rptr := &reporters.Chain{[]miners.Reporter{reporters.NewLog(fmtr, false, "DEBUG", "common-ancestor"), uniq}}

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

