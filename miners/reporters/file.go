package reporters

import (
	"fmt"
	"io"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type File struct {
	config     *config.Config
	fmtr        lattice.Formatter
	prfmtr      lattice.PrFormatter
	patterns   io.WriteCloser
	embeddings io.WriteCloser
	names      io.WriteCloser
	matrices   io.WriteCloser
	prs        io.WriteCloser
}

func NewFile(c *config.Config, fmtr lattice.Formatter, showPr bool, patternsFilename, embeddingsFilename, namesFilename, matricesFilename, prsFilename string) (*File, error) {
	patterns, err := os.Create(c.OutputFile(patternsFilename + fmtr.FileExt()))
	if err != nil {
		return nil, err
	}
	embeddings, err := os.Create(c.OutputFile(embeddingsFilename + fmtr.FileExt()))
	if err != nil {
		return nil, err
	}
	names, err := os.Create(c.OutputFile(namesFilename))
	if err != nil {
		return nil, err
	}
	var matrices io.WriteCloser
	var prs io.WriteCloser
	var prfmtr lattice.PrFormatter
	if showPr {
		prfmtr = fmtr.PrFormatter()
		if prfmtr != nil {
			prs, err = os.Create(c.OutputFile(prsFilename))
			if err != nil {
				return nil, err
			}
			matrices, err = os.Create(c.OutputFile(matricesFilename))
			if err != nil {
				return nil, err
			}
		}
	}
	r := &File{
		config:     c,
		fmtr:        fmtr,
		prfmtr:      prfmtr,
		patterns:   patterns,
		embeddings: embeddings,
		names:      names,
		matrices:   matrices,
		prs:        prs,
	}
	return r, nil
}

func (r *File) Report(n lattice.Node) error {
	err := r.fmtr.FormatPattern(r.patterns, n)
	if err != nil {
		return err
	}
	err = r.fmtr.FormatEmbeddings(r.embeddings, n)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.names, "%v\n", r.fmtr.PatternName(n))
	if err != nil {
		return err
	}
	if r.prfmtr != nil {
		matrices, err := r.prfmtr.Matrices(n)
		if err == nil {
			r.prfmtr.FormatMatrices(r.matrices, r.fmtr, n, matrices)
		}
		if err != nil {
			fmt.Fprintf(r.matrices, "ERR: %v\n", err)
			errors.Logf("ERROR", "Pr Matrices Computation Error: vs", err)
		} else if r.prfmtr.CanComputeSelPr(n, matrices) {
			pr, err := r.prfmtr.SelectionProbability(n, matrices)
			if err != nil {
				fmt.Fprintf(r.prs, "ERR: %v\n", err)
				errors.Logf("ERROR", "PrComputation Error: %v", err)
			} else {
				fmt.Fprintf(r.prs, "%g, %v\n", pr, r.fmtr.PatternName(n))
			}
		} else {
			fmt.Fprintf(r.prs, "SKIPPED\n")
		}
	}
	return nil
}

func (r *File) Close() error {
	err := r.patterns.Close()
	if err != nil {
		return err
	}
	err = r.embeddings.Close()
	if err != nil {
		return err
	}
	return nil
}
