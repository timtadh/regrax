package absorbing

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/matrix"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/reporters"
	"github.com/timtadh/sfp/miners/walker"
)

type SparseEntry struct {
	Row, Col int
	Value    float64
	Inverse  int
}

type Sparse struct {
	Rows, Cols int
	Entries    []SparseEntry
}

type Walker struct {
	walker.Walker
	Errors chan error
}

func NewWalker(conf *config.Config) *Walker {
	miner := &Walker{
		Errors: make(chan error),
	}
	miner.Walker = *walker.NewWalker(conf, MakeAbsorbingWalk(MakeSample(miner), miner.Errors))
	return miner
}

func (w *Walker) PrMatrices(n lattice.Node) (Q, R, u *Sparse, err error) {
	lat, err := lattice.MakeLattice(n)
	if err != nil {
		return nil, nil, nil, err
	}
	p, err := w.probabilities(lat)
	if err != nil {
		return nil, nil, nil, err
	}
	Q = &Sparse{
		Rows:    len(lat.V) - 1,
		Cols:    len(lat.V) - 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	R = &Sparse{
		Rows:    len(lat.V) - 1,
		Cols:    1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	u = &Sparse{
		Rows:    1,
		Cols:    len(lat.V) - 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	for i, x := range lat.V {
		if c, err := x.ParentCount(); err != nil {
			return nil, nil, nil, err
		} else if c == 0 {
			u.Entries = append(u.Entries, SparseEntry{0, i, 1.0, 1})
		}
	}
	for _, e := range lat.E {
		if e.Targ >= len(lat.V)-1 {
			R.Entries = append(R.Entries, SparseEntry{e.Src, 0, 1.0 / float64(p[e.Src]), p[e.Src]})
		} else {
			Q.Entries = append(Q.Entries, SparseEntry{e.Src, e.Targ, 1.0 / float64(p[e.Src]), p[e.Src]})
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

func (w *Walker) SelectionProbability(Q_, R_, u_ *Sparse) (float64, error) {
	if Q_.Rows == 0 && Q_.Cols == 0 {
		return 1.0 / float64(len(w.Start)), nil
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

func (w *Walker) probabilities(lat *lattice.Lattice) ([]int, error) {
	P := make([]int, len(lat.V))
	for i, node := range lat.V {
		count, err := node.ChildCount()
		if err != nil {
			return nil, err
		}
		if i+1 == len(lat.V) {
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

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter) error {
	mr, err := NewMatrixReporter(w, w.Errors)
	if err != nil {
		return err
	}
	return (w.Walker).Mine(dt, &reporters.Chain{[]miners.Reporter{rptr, mr}})
}

func MakeAbsorbingWalk(sample func(lattice.Node) (lattice.Node, error), errs chan error) walker.Walk {
	return func(wlkr *walker.Walker) (chan lattice.Node, chan bool, chan error) {
		samples := make(chan lattice.Node)
		terminate := make(chan bool)
		go func() {
		loop:
			for {
				sampled, err := sample(wlkr.Dt.Empty())
				if err != nil {
					errs <- err
					break loop
				}
				select {
				case <-terminate:
					break loop
				case samples <- sampled:
				}
			}
			close(samples)
			close(errs)
			close(terminate)
		}()
		return samples, terminate, errs
	}
}

type Transitioner interface {
	Next(cur lattice.Node) (lattice.Node, error)
}

func MakeSample(t Transitioner) func(lattice.Node) (lattice.Node, error) {
	return func(empty lattice.Node) (max lattice.Node, err error) {
		cur := empty
		next, err := t.Next(cur)
		if err != nil {
			return nil, err
		}
		for next != nil {
			cur = next
			next, err = t.Next(cur)
			// errors.Logf("DEBUG", "compute next %v %v", next, err)
			if err != nil {
				return nil, err
			}
			// errors.Logf("DEBUG", "cur %v kids %v next %v", cur, kidCount, next)
		}
		return cur, nil
	}
}

func (w *Walker) Next(cur lattice.Node) (lattice.Node, error) {
	return uniform(cur.Children())
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
