package uniprox

import (
	"io"
)

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)

type PrFormatter struct {
	w        *Walker
}

func NewPrFormatter(w *Walker) *PrFormatter {
	return &PrFormatter{
		w: w,
	}
}

func (r *PrFormatter) Matrices(n lattice.Node) (interface{}, error) {
	return nil, nil
}

func (r *PrFormatter) CanComputeSelPr(n lattice.Node, m interface{}) bool {
	return true
}

func (r *PrFormatter) SelectionProbability(n lattice.Node, m interface{}) (float64, error) {
	var nPr float64
	err := r.w.Prs.DoFind(n.Pattern().Label(), func(_ []byte, npr float64) error {
		nPr = npr
		return nil
	})
	if err != nil {
		return 0, err
	}
	return nPr, nil
}

func (r *PrFormatter) FormatMatrices(w io.Writer, fmtr lattice.Formatter, n lattice.Node, m interface{}) (error) {
	return nil
}

