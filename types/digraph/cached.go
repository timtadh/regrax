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
		node, err := LoadEmbListNode(dt, adj)
		if err != nil {
			return err
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
