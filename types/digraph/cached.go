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

type Saveable interface {
	Save() error
}

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
		case Saveable:
			err = node.Save()
			if err != nil {
				return err
			}
		default:
			return errors.Errorf("unexpected lattice.Node type %T %v", n, n)
		}
		err = cache.Add(key, n.Pattern().Label())
		if err != nil {
			return err
		}
	}
	return nil
}

func cached(n Node, dt *Digraph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap) (nodes []lattice.Node, has bool, err error) {
	key := n.Label()
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
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
	return nodes, true, nil
}
