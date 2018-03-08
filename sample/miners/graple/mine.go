package graple

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/matrix"
)

import (
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/lattice"
	"github.com/timtadh/regrax/sample/miners/walker"
)

type Matrices struct {
	Q, R, U *Sparse
}

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

func (w *Walker) PrFormatter() lattice.PrFormatter {
	return NewPrFormatter(w)
}

func (w *Walker) PrMatrices(n lattice.Node) (QRu *Matrices, err error) {
	lat, err := lattice.MakeLattice(n)
	if err != nil {
		return nil, err
	}
	p, err := w.probabilities(lat)
	if err != nil {
		return nil, err
	}
	QRu = new(Matrices)
	QRu.Q = &Sparse{
		Rows:    len(lat.V) - 1,
		Cols:    len(lat.V) - 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	QRu.R = &Sparse{
		Rows:    len(lat.V) - 1,
		Cols:    1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	QRu.U = &Sparse{
		Rows:    1,
		Cols:    len(lat.V) - 1,
		Entries: make([]SparseEntry, 0, len(lat.V)-1),
	}
	for i, x := range lat.V {
		if c, err := x.ParentCount(); err != nil {
			return nil, err
		} else if c == 0 {
			QRu.U.Entries = append(QRu.U.Entries, SparseEntry{0, i, 1.0, 1})
		}
	}
	for _, e := range lat.E {
		if e.Targ >= len(lat.V)-1 {
			QRu.R.Entries = append(QRu.R.Entries, SparseEntry{e.Src, 0, 1.0 / float64(p[e.Src]), p[e.Src]})
		} else {
			QRu.Q.Entries = append(QRu.Q.Entries, SparseEntry{e.Src, e.Targ, 1.0 / float64(p[e.Src]), p[e.Src]})
		}
	}
	return QRu, nil
}

func (m *Sparse) Dense() *matrix.DenseMatrix {
	d := matrix.Zeros(m.Rows, m.Cols)
	for _, e := range m.Entries {
		d.Set(e.Row, e.Col, e.Value)
	}
	return d
}

func (w *Walker) SelectionProbability(Q_, R_, u_ *Sparse) (float64, error) {
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
		if count < len(lat.Children(i)) {
			errors.Logf("WARNING", "count < lat.Children count %v < %v", count, len(lat.Children(i)))
			count = len(lat.Children(i))
		}
		if count > 0 && len(lat.Children(i)) == 0 {
			errors.Logf("WARNING", "count > 0 && lat.Children == 0 : %v > 0 lat.Children == %v", count, len(lat.Children(i)))
		}
		if i+1 == len(lat.V) {
			P[i] = -1
		} else if count == 0 {
			P[i] = 1
			errors.Logf("WARNING", "0 count for %v, using %v", node, P[i])
		} else {
			P[i] = count
		}
	}
	return P, nil
}

func MakeAbsorbingWalk(sample func(lattice.Node) (lattice.Node, error), errs chan error) walker.Walk {
	return func(wlkr *walker.Walker) (chan lattice.Node, chan bool, chan error) {
		samples := make(chan lattice.Node)
		terminate := make(chan bool)
		go func() {
			for {
				sampled, err := sample(wlkr.Dt.Root())
				if err != nil {
					errs <- err
					break
				}
				samples <- sampled
				if <-terminate {
					break
				}
			}
			close(samples)
			close(errs)
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
	errors.Logf("DEBUG", "cur %v", cur)
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
