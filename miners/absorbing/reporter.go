package absorbing

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type MatrixReporter struct {
	w        *Walker
	matrices io.WriteCloser
	prs      io.WriteCloser
	reports  chan lattice.Node
	errors   chan error
}

func NewMatrixReporter(w *Walker, errors chan error) (*MatrixReporter, error) {
	matrices, err := os.Create(w.Config.OutputFile("pr-matrices.jsons"))
	if err != nil {
		return nil, err
	}
	prs, err := os.Create(w.Config.OutputFile("selection-probabilities.prs"))
	if err != nil {
		return nil, err
	}
	r := &MatrixReporter{
		w:        w,
		matrices: matrices,
		prs:      prs,
		reports:  make(chan lattice.Node),
		errors:   errors,
	}
	go r.processReports()
	return r, nil
}

func (r *MatrixReporter) Report(n lattice.Node) error {
	r.reports <- n
	return nil
}

func (r *MatrixReporter) processReports() {
	r.w.Config.AsyncTasks.Add(1)
	for n := range r.reports {
		err := r.report(n)
		if err != nil {
			errors.Logf("ERROR", "%v", err)
			r.errors <- err
			break
		}
	}
	r.w.Config.AsyncTasks.Done()
}

func (r *MatrixReporter) report(n lattice.Node) error {
	Q, R, u, err := r.w.PrMatrices(n)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(map[string]interface{}{
		"Q":              Q,
		"R":              R,
		"u":              u,
		"startingPoints": 1,
	})
	if err != nil {
		return err
	}
	_, err = r.matrices.Write(bytes)
	if err != nil {
		return err
	}
	_, err = r.matrices.Write([]byte("\n"))
	if err != nil {
		return err
	}
	if Q.Cols > 1000 {
		_, err = fmt.Fprintf(r.prs, "NA\n")
		return err
	} else {
		pr, err := r.w.SelectionProbability(Q, R, u)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(r.prs, "%v\n", pr)
		return err
	}
}

func (r *MatrixReporter) Close() error {
	close(r.reports)
	r.w.Config.AsyncTasks.Wait()
	err := r.matrices.Close()
	if err != nil {
		return err
	}
	err = r.prs.Close()
	if err != nil {
		return err
	}
	return nil
}
