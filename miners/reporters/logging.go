package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
)


type LoggingReporter struct {
	Next miners.Reporter
}

func (lr *LoggingReporter) Report(n lattice.Node) error {
	errors.Logf("INFO", "sample %v", n)
	if lr.Next != nil {
		return lr.Next.Report(n)
	}
	return nil
}

func (lr *LoggingReporter) Close() error {
	if lr.Next != nil {
		return lr.Next.Close()
	}
	return nil
}
