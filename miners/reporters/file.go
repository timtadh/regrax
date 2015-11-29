package reporters

import (
	"fmt"
	"io"
	"os"
)

import (
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)


type FileReporter struct {
	config *config.Config
	fmt lattice.Formatter
	patterns io.WriteCloser
	embeddings io.WriteCloser
}


func NewFileReporter(c *config.Config, fmt lattice.Formatter) (*FileReporter, error) {
	patterns, err := os.Create(c.OutputFile("patterns" + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	embeddings, err := os.Create(c.OutputFile("embeddings" + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	r := &FileReporter{
		config: c,
		fmt: fmt,
		patterns: patterns,
		embeddings: embeddings,
	}
	return r, nil
}

func (r *FileReporter) Report(n lattice.Node) error {
	name := r.fmt.PatternName(n)
	_, err := fmt.Fprintf(r.patterns, "\\\\ %s\n\n%s\n", name, r.fmt.FormatPattern(n))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.embeddings, "\\\\ %s\n\n", name)
	if err != nil {
		return err
	}
	for _, embedding := range r.fmt.FormatEmbeddings(n) {
		_, err = fmt.Fprintf(r.embeddings, "%s\n", embedding)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *FileReporter) Close() error {
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


