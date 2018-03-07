package graple

import (
	"encoding/json"
	"io"
)

import ()

import (
	"github.com/timtadh/regrax/lattice"
)

type PrFormatter struct {
	w *Walker
}

func NewPrFormatter(w *Walker) *PrFormatter {
	return &PrFormatter{
		w: w,
	}
}

func (r *PrFormatter) Matrices(n lattice.Node) (interface{}, error) {
	return r.w.PrMatrices(n)
}

func (r *PrFormatter) CanComputeSelPr(n lattice.Node, m interface{}) bool {
	QRu := m.(*Matrices)
	if QRu.Q.Cols > 200 {
		return false
	}
	return true
}

func (r *PrFormatter) SelectionProbability(n lattice.Node, m interface{}) (float64, error) {
	QRu := m.(*Matrices)
	return r.w.SelectionProbability(QRu.Q, QRu.R, QRu.U)
}

func (r *PrFormatter) FormatMatrices(w io.Writer, fmtr lattice.Formatter, n lattice.Node, m interface{}) error {
	QRu := m.(*Matrices)
	bytes, err := json.Marshal(map[string]interface{}{
		"Name":           fmtr.PatternName(n),
		"Q":              QRu.Q,
		"R":              QRu.R,
		"u":              QRu.U,
		"startingPoints": 1,
	})
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}
