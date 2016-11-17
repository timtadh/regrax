package reporters

import (
	"fmt"
	"os"
	"path/filepath"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type Dir struct {
	config *config.Config
	fmtr    lattice.Formatter
	prfmtr  lattice.PrFormatter
	dir    string
	count  int
}

func NewDir(c *config.Config, fmtr lattice.Formatter, showPr bool, dirname string) (*Dir, error) {
	samples := c.OutputFile(dirname)
	err := os.MkdirAll(samples, 0775)
	if err != nil {
		return nil, err
	}
	var prfmtr lattice.PrFormatter
	if showPr {
		prfmtr = fmtr.PrFormatter()
	}
	r := &Dir{
		config: c,
		fmtr:    fmtr,
		prfmtr:  prfmtr,
		dir:    samples,
	}
	return r, nil
}

func (r *Dir) Report(n lattice.Node) error {
	dir := filepath.Join(r.dir, fmt.Sprintf("%d", r.count))
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}
	r.count++
	name, err := os.Create(filepath.Join(dir, "pattern.name"))
	if err != nil {
		return err
	}
	defer name.Close()
	fmt.Fprintf(name, "%s\n", r.fmtr.PatternName(n))
	pattern, err := os.Create(filepath.Join(dir, "pattern"+r.fmtr.FileExt()))
	if err != nil {
		return err
	}
	defer pattern.Close()
	err = r.fmtr.FormatPattern(pattern, n)
	if err != nil {
		return err
	}
	embs, err := r.fmtr.Embeddings(n)
	if err != nil {
		return err
	}
	count, err := os.Create(filepath.Join(dir, "embeddings"))
	if err != nil {
		return err
	}
	defer count.Close()
	fmt.Fprintf(count, "%d\n", len(embs))
	for i, emb := range embs {
		edir := filepath.Join(dir, fmt.Sprintf("%d", i))
		err := os.MkdirAll(edir, 0775)
		if err != nil {
			return err
		}
		embedding, err := os.Create(filepath.Join(edir, "embedding"+r.fmtr.FileExt()))
		if err != nil {
			return err
		}
		defer embedding.Close()
		fmt.Fprintf(embedding, "%s\n", emb)
	}
	if r.prfmtr != nil {
		matrices, err := r.prfmtr.Matrices(n)
		if err == nil {
			mw, err := os.Create(filepath.Join(dir, "matrices.json"))
			if err != nil {
				return err
			}
			defer mw.Close()
			r.prfmtr.FormatMatrices(mw, r.fmtr, n, matrices)
		}
		if err != nil {
			errors.Logf("ERROR", "Pr Matrices Computation Error: %v", err)
		} else if r.prfmtr.CanComputeSelPr(n, matrices) {
			pr, err := r.prfmtr.SelectionProbability(n, matrices)
			if err != nil {
				errors.Logf("ERROR", "PrComputation Error: %v", err)
			} else {
				pw, err := os.Create(filepath.Join(dir, "pattern.pr"))
				if err != nil {
					return err
				}
				defer pw.Close()
				fmt.Fprintf(pw, "%g\n", pr)
			}
		}
	}
	return nil
}

func (r *Dir) Close() error {
	count, err := os.Create(filepath.Join(r.dir, "count"))
	if err != nil {
		return err
	}
	defer count.Close()
	fmt.Fprintf(count, "%d\n", r.count)
	return nil
}
