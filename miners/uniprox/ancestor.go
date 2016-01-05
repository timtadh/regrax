package uniprox

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/graph"
	"github.com/timtadh/sfp/types/itemset"
)

func CommonAncestor(patterns []lattice.Pattern) (_ lattice.Pattern, err error) {
	if len(patterns) == 0 {
		return nil, errors.Errorf("no patterns given")
	}
	switch patterns[0].(type) {
	case *graph.Pattern: return graphCommonAncestor(patterns)
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

func graphCommonAncestor(patterns []lattice.Pattern) (lattice.Pattern, error) {
	return nil, errors.Errorf("unimplemented")
}

