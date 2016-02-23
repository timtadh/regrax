package digraph

import (
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_int"
)

func cache(dt *Graph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte, nodes []lattice.Node) (err error) {
	if has, err := count.Has(key); err != nil {
		return err
	} else if has {
		return nil
	}
	err = count.Add(key, int32(len(nodes)))
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if dt.search {
			return errors.Errorf("unimplemented")
		} else {
			err = node.(*Node).Save()
			if err != nil {
				return err
			}
			err = cache.Add(key, node.(*Node).pat.label)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cached(dt *Graph, count bytes_int.MultiMap, cache bytes_bytes.MultiMap, key []byte) (nodes []lattice.Node, has bool, err error) {
	if has, err := count.Has(key); err != nil {
		return nil, false, err
	} else if !has {
		return nil, false, nil
	}
	err = cache.DoFind(key, func(_, adj []byte) error {
		if dt.search {
			return errors.Errorf("unimplemented")
		} else {
			sgs := make(SubGraphs, 0, 10)
			err := dt.Embeddings.DoFind(adj, func(_ []byte, sg *goiso.SubGraph) error {
				sgs = append(sgs, sg)
				return nil
			})
			if err != nil {
				return err
			}
			nodes = append(nodes, &Node{Pattern{label: adj}, dt, sgs})
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return nodes, true, nil
}


