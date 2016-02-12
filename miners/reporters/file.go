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
	fmt        lattice.Formatter
	prfmt      lattice.PrFormatter
	patterns   io.WriteCloser
	embeddings io.WriteCloser
	matrices   io.WriteCloser
	prs        io.WriteCloser
}

func NewFile(c *config.Config, fmt lattice.Formatter, showPr bool, patternsFilename, embeddingsFilename, matricesFilename, prsFilename string) (*File, error) {
	patterns, err := os.Create(c.OutputFile(patternsFilename + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	embeddings, err := os.Create(c.OutputFile(embeddingsFilename + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	var matrices io.WriteCloser
	var prs io.WriteCloser
	var prfmt lattice.PrFormatter
	if showPr {
		prfmt = fmt.PrFormatter()
		prs, err = os.Create(c.OutputFile(prsFilename))
		if err != nil {
			return nil, err
		}
		matrices, err = os.Create(c.OutputFile(matricesFilename))
		if err != nil {
			return nil, err
		}
	}
	r := &File{
		config:     c,
		fmt:        fmt,
		prfmt:      prfmt,
		patterns:   patterns,
		embeddings: embeddings,
		matrices:   matrices,
		prs:        prs,
	}
	return r, nil
}

func (r *File) Report(n lattice.Node) error {
	err := r.fmt.FormatPattern(r.patterns, n)
	if err != nil {
		return err
	}
	err = r.fmt.FormatEmbeddings(r.embeddings, n)
	if err != nil {
		return err
	}
	if r.prfmt != nil {
		matrices, err := r.prfmt.Matrices(n)
		if err == nil {
			r.prfmt.FormatMatrices(r.matrices, n, matrices)
		} else if err != nil {
			fmt.Fprintf(r.matrices, "ERR: %v\n", err)
			errors.Logf("ERROR", "Pr Matrices Computation Error: vs", err)
		} else if r.prfmt.CanComputeSelPr(n, matrices) {
			pr, err := r.prfmt.SelectionProbability(n, matrices)
			if err != nil {
				fmt.Fprintf(r.prs, "ERR: %v\n", err)
				errors.Logf("ERROR", "PrComputation Error: %v", err)
			} else {
				fmt.Fprintf(r.prs, "%g\n", pr)
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
