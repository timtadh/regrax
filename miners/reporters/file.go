package reporters

import (
	"io"
	"os"
)

import ()

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type File struct {
	config     *config.Config
	fmt        lattice.Formatter
	patterns   io.WriteCloser
	embeddings io.WriteCloser
}

func NewFile(c *config.Config, fmt lattice.Formatter, patternsFilename, embeddingsFilename string) (*File, error) {
	patterns, err := os.Create(c.OutputFile(patternsFilename + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	embeddings, err := os.Create(c.OutputFile(embeddingsFilename + fmt.FileExt()))
	if err != nil {
		return nil, err
	}
	r := &File{
		config:     c,
		fmt:        fmt,
		patterns:   patterns,
		embeddings: embeddings,
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
