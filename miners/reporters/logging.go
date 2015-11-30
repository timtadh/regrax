package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Log struct {
	count int
}

func (lr *Log) Report(n lattice.Node) error {
	lr.count++
	errors.Logf("INFO", "sample %v %v", lr.count, n)
	return nil
}

func (lr *Log) Close() error {
	return nil
}
