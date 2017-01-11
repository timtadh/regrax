package reporters

import (
	"fmt"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)

type Count struct {
	config     *config.Config
	count      int
	filename   string
}

func NewCount(c *config.Config, filename string) (*Count, error) {
	r := &Count{
		config:    c,
		filename:  filename,
	}
	return r, nil
}

func (r *Count) Report(n lattice.Node) error {
	r.count++
	return nil
}

func (r *Count) Close() error {
	errors.Logf("INFO", "total graphs found %v", r.count)
	f, err := os.Create(r.config.OutputFile(r.filename))
	if err != nil {
		return err
	}
	_, perr := fmt.Fprintf(f, "%v\n", r.count)
	err = f.Close()
	if perr != nil {
		return perr
	}
	if err != nil {
		return err
	}
	return nil
}
