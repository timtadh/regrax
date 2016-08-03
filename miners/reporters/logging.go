package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Log struct {
	fmtr    lattice.Formatter
	prs    bool
	level  string
	prefix string
	count  int
}

func NewLog(fmtr lattice.Formatter, prs bool, level, prefix string) *Log {
	if level == "" {
		level = "INFO"
	}
	return &Log{fmtr: fmtr, prs: prs, level: level, prefix: prefix}
}

func (lr *Log) Report(n lattice.Node) error {
	lr.count++
	prfmtr := lr.fmtr.PrFormatter()
	pr := -1.0
	if lr.prs && prfmtr != nil {
		matrices, err := prfmtr.Matrices(n)
		if err != nil {
			errors.Logf("ERROR", "Pr Matrices Computation Error: %v", err)
		} else if prfmtr.CanComputeSelPr(n, matrices) {
			pr, err = prfmtr.SelectionProbability(n, matrices)
			if err != nil {
				errors.Logf("ERROR", "PrComputation Error: %v", err)
			}
		}
	}
	if lr.prefix != "" && pr > -1.0 {
		errors.Logf(lr.level, "%s %v (pr = %5.3g) %v", lr.prefix, lr.count, pr, n)
	} else if lr.prefix != "" {
		errors.Logf(lr.level, "%s %v %v", lr.prefix, lr.count, n)
	} else if pr > -1.0 {
		errors.Logf(lr.level, "%v (pr = %5.3g) %v", lr.count, pr, n)
	} else {
		errors.Logf(lr.level, "%v %v", lr.count, n)
	}
	return nil
}

func (lr *Log) Close() error {
	return nil
}
