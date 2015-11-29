package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)


type LoggingReporter struct {
}

func (lr *LoggingReporter) Report(n lattice.Node) error {
	errors.Logf("INFO", "sample %v", n)
	return nil
}

func (lr *LoggingReporter) Close() error {
	return nil
}

