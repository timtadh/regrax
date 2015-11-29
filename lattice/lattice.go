package lattice

import (
)

func MakeLattice(n Node) (*Lattice, error) {
	lat, err := n.Lattice()
	if err != nil {
		_, ok := err.(*NoLattice)
		if !ok {
			return nil, err
		}
	} else {
		return lat, nil
	}
	return lattice(n)
}

func lattice(node Node) (*Lattice, error) {
	pop := func(queue []Node) (Node, []Node) {
		n := queue[0]
		copy(queue[0:len(queue)-1],queue[1:len(queue)])
		queue = queue[0:len(queue)-1]
		return n, queue
	}
	queue := make([]Node, 0, 10)
	queue = append(queue, node)
	queued := make(map[string]bool)
	rlattice := make([]Node, 0, 10)
	for len(queue) > 0 {
		var n Node
		n, queue = pop(queue)
		queued[string(n.Label())] = true
		rlattice = append(rlattice, n)
		parents, err := n.Parents()
		if err != nil {
			return nil, err
		}
		for _, p := range parents {
			l := string(p.Label())
			if _, has := queued[l]; !has {
				queue = append(queue, p)
				queued[l] = true
			}
		}
	}
	lattice := make([]Node, 0, len(rlattice))
	labels := make(map[string]int,len(lattice))
	for i := len(rlattice)-1; i >= 0; i-- {
		lattice = append(lattice, rlattice[i])
		labels[string(lattice[len(lattice)-1].Label())] = len(lattice)-1
	}
	edges := make([]Edge, 0, len(lattice)*2)
	for i, n := range lattice {
		kids, err := n.Children()
		if err != nil {
			return nil, err
		}
		for _, kid := range kids {
			j, has := labels[string(kid.Label())]
			if has {
				edges = append(edges, Edge{Src: i, Targ: j})
			}
		}
	}
	return &Lattice{lattice, edges}, nil
}
