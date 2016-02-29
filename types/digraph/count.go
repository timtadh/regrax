package digraph

import (
)

import (
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/stores/bytes_int"
)


func count(n Node, compute func()([]lattice.Node, error), counts bytes_int.MultiMap) (int, error) {
	if has, err := counts.Has(n.Label()); err != nil {
		return 0, err
	} else if !has {
		nodes, err := compute()
		if err != nil {
			return 0, err
		}
		return len(nodes), nil
	}
	var count int32
	err := counts.DoFind(n.Label(), func(_ []byte, c int32) error {
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

