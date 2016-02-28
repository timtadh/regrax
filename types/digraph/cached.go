package digraph

import (
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

func cache(dt *Digraph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte, nodes []lattice.Node) (err error) {
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
		switch node := n.(type) {
		case *SearchNode:
			return errors.Errorf("unimplemented")
		case *EmbListNode:
			err = node.Save()
			if err != nil {
				return err
			}
			err = cache.Add(key, node.pat.label)
			if err != nil {
				return err
			}
		default:
			return errors.Errorf("unexpected lattice.Node type %T %v", n, n)
		}
	}
	return nil
}

func cached(dt *Digraph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte) (nodes []lattice.Node, has bool, err error) {
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
	err = cache.DoFind(key, func(_, adj []byte) error {
		if dt.search {
			return errors.Errorf("unimplemented")
		} else {
			node, err := LoadEmbListNode(dt, adj)
			if err != nil {
				return err
			}
			nodes = append(nodes, node)
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return nodes, true, nil
}


