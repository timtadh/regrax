package digraph

import ()

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

func cacheAdj(dt *Digraph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte, nodes []lattice.Node) (err error) {
	dt.lock.Lock()
	defer dt.lock.Unlock()
	if false {
		pat, _ := LoadSubgraphPattern(dt, key)
		errors.Logf("WARN", "skipped caching %v", pat.Pat)
		return nil
	}
	if has, err := count.Has(key); err != nil {
		return err
	} else if has {
		return nil
	}
	err = count.Add(key, int32(len(nodes)))
	if err != nil {
		return err
	}
	for _, n := range nodes {
		err = cache.Add(key, n.Pattern().Label())
		if err != nil {
			return err
		}
	}
	return nil
}

func cachedAdj(n Node, dt *Digraph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap) (nodes []lattice.Node, has bool, err error) {
	dt.lock.RLock()
	defer dt.lock.RUnlock()
	key := n.Label()
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
	// errors.Logf("DEBUG", "loading %v", n)
	err = cache.DoFind(key, func(_, adj []byte) (err error) {
		// WHY DO WE NEED TO UNLOCK?
		// We will aquire this lock in the course of LoadEmbList in READ MODE
		// This is fine to re-aquire in READ
		// However, if another thread tries to aquire in WRITE before we re-aquire
		// There will be a waiting WRITE when we try to aquire READ
		// A waiting WRITE will prevent all READ aquisitions
		// DEADLOCK.
		dt.lock.RUnlock()
		defer dt.lock.RLock()
		node, err := LoadEmbListNode(dt, adj)
		if err != nil {
			return err
		}
		if node == nil {
			errors.Logf("ERROR", "node is nil")
			panic("node was nil")
		}
		nodes = append(nodes, node)
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if false {
		pat, _ := LoadSubgraphPattern(dt, key)
		errors.Logf("LOAD-DEBUG", "Loaded Cached Adj %v adj %v", pat.Pat, len(nodes))
	}
	return nodes, true, nil
}
