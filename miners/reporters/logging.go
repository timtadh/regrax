package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)


type Log struct {
}

func (lr *Log) Report(n lattice.Node) error {
	errors.Logf("INFO", "sample %v", n)
	return nil
}

func (lr *Log) Close() error {
	return nil
}

