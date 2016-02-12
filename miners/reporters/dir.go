package reporters

import (
	"fmt"
	"os"
	"path/filepath"
)

import ()

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type Dir struct {
	config *config.Config
	fmt    lattice.Formatter
	dir    string
	count  int
}

func NewDir(c *config.Config, fmt lattice.Formatter, dirname string) (*Dir, error) {
	samples := c.OutputFile(dirname)
	err := os.MkdirAll(samples, 0775)
	if err != nil {
		return nil, err
	}
	r := &Dir{
		config: c,
		fmt:    fmt,
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
	fmt.Fprintf(name, "%s\n", r.fmt.PatternName(n))
	pattern, err := os.Create(filepath.Join(dir, "pattern" + r.fmt.FileExt()))
	if err != nil {
		return err
	}
	defer pattern.Close()
	err = r.fmt.FormatPattern(pattern, n)
	if err != nil {
		return err
	}
	embs, err := r.fmt.Embeddings(n)
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
		embedding, err := os.Create(filepath.Join(edir, "embedding" + r.fmt.FileExt()))
		if err != nil {
			return err
		}
		defer embedding.Close()
		fmt.Fprintf(embedding, "%s\n", emb)
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
