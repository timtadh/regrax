package absorbing

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/matrix"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners/walker"
)


type SparseEntry struct {
	Row, Col int
	Value float64
	Inverse int
}

type Sparse struct {
	Rows, Cols int
	Entries []SparseEntry
}

/*
func (m *Miner) Mine(input lattice.Input, dt lattice.DataType) error {
	err := m.init(input, dt)
	if err != nil {
		return err
	}
	for s := 0; s < m.config.Samples; s++ {
		sampled, err := m.rejectingWalk(dt)
		if err != nil {
			return err
		}
		errors.Logf("INFO", "sample %v %v", sampled, sampled.Label())
		Q, R, u, err := m.PrMatrices(sampled, dt)
		if err != nil {
			return err
		}
		errors.Logf("INFO", "matrix Q %v", Q)
		errors.Logf("INFO", "matrix R %v", R)
		errors.Logf("INFO", "matrix u %v", u)
		pr, err := m.SelectionProbability(Q, R, u)
		if err != nil {
			return err
		}
		errors.Logf("INFO", "sel pr %v", pr)
		errors.Logf("INFO", "")
	}
	return nil
}*/

func PrMatrices(w *walker.Walker, n lattice.Node) (Q, R, u *Sparse, err error) {
	lat, err := lattice.MakeLattice(n, w.Config.Support, w.Dt)
	if err != nil {
		return nil, nil, nil, err
	}
	p, err := probabilities(w, lat)
	if err != nil {
		return nil, nil, nil, err
	}
	Q = &Sparse{
		Rows: len(lat.V)-1,
		Cols: len(lat.V)-1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	R = &Sparse{
		Rows: len(lat.V)-1,
		Cols: 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	u = &Sparse{
		Rows: 1,
		Cols: len(lat.V)-1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	sp := len(w.Start)
	for i, x := range lat.V {
		if x.StartingPoint() && i < len(lat.V)-1 {
			u.Entries = append(u.Entries, SparseEntry{0, i, 1.0/float64(sp), sp})
		}
	}
	for _, e := range lat.E {
		if e.Targ >= len(lat.V)-1 {
			R.Entries = append(R.Entries, SparseEntry{e.Src, 0, 1.0/float64(p[e.Src]), p[e.Src]})
		} else {
			Q.Entries = append(Q.Entries, SparseEntry{e.Src, e.Targ, 1.0/float64(p[e.Src]), p[e.Src]})
		}
	}
	return Q, R, u, nil
}

func (m *Sparse) Dense() *matrix.DenseMatrix {
	d := matrix.Zeros(m.Rows, m.Cols)
	for _, e := range m.Entries {
		d.Set(e.Row, e.Col, e.Value)
	}
	return d
}

func SelectionProbability(w *walker.Walker, Q_, R_, u_ *Sparse) (float64, error) {
	if Q_.Rows == 0 && Q_.Cols == 0 {
		return 1.0/float64(len(w.Start)), nil
	}
	Q := Q_.Dense()
	R := R_.Dense()
	u := u_.Dense()
	I := matrix.Eye(Q.Rows())
	IQ, err := I.Minus(Q)
	if err != nil {
		return 0, err
	}
	N := matrix.Inverse(IQ)
	B, err := N.Times(R)
	if err != nil {
		return 0, err
	}
	P, err := u.Times(B)
	if err != nil {
		return 0, err
	}
	if P.Rows() != P.Cols() && P.Rows() != 1 {
		return 0, errors.Errorf("Unexpected P shape %v %v", P.Rows(), P.Cols())
	}
	x := P.Get(0, 0)
	if x > 1.0 || x != x {
		return 0, errors.Errorf("could not accurately compute p")
	}
	return x, nil
}

func probabilities(w *walker.Walker, lat *lattice.Lattice) ([]int, error) {
	P := make([]int, len(lat.V))
	for i, node := range lat.V {
		count, err := node.ChildCount(w.Config.Support, w.Dt)
		if err != nil {
			return nil, err
		}
		if i + 1 == len(lat.V) {
			P[i] = -1
		} else if count == 0 {
			P[i] = 1
			errors.Logf("INFO", "0 count for %v", node)
		} else {
			P[i] = count
		}
	}
	return P, nil
}

func RejectingWalk(w *walker.Walker) (chan lattice.Node, chan error) {
	nodes := make(chan lattice.Node)
	errors := make(chan error)
	go func() {
		i := 0
		for i < w.Config.Samples {
			if sampled, err := walk(w); err != nil {
				errors<-err
				break
			} else if sampled.Size() >= w.Config.MinSize {
				nodes<-sampled
				i++
			}
		}
		close(nodes)
		close(errors)
	}()
	return nodes, errors
}

func walk(w *walker.Walker) (max lattice.Node, err error) {
	cur, _ := uniform(w.Start, nil)
	// errors.Logf("DEBUG", "start %v", cur)
	next, err := uniform(cur.Children(w.Config.Support, w.Dt))
	if err != nil {
		return nil, err
	}
	for next != nil {
		cur = next
		next, err = uniform(cur.Children(w.Config.Support, w.Dt))
		// errors.Logf("DEBUG", "compute next %v %v", next, err)
		if err != nil {
			return nil, err
		}
		if next != nil && next.Size() > w.Config.MaxSize {
			next = nil
		}
		// errors.Logf("DEBUG", "cur %v kids %v next %v", cur, kidCount, next)
	}
	return cur, nil
}

func uniform(slice []lattice.Node, err error) (lattice.Node, error) {
	// errors.Logf("DEBUG", "children %v", slice)
	if err != nil {
		return nil, err
	}
	if len(slice) > 0 {
		return slice[rand.Intn(len(slice))], nil
	}
	return nil, nil
}

