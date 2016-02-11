package reporters

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Log struct {
	level  string
	prefix string
	count  int
}

func NewLog(level, prefix string) *Log {
	if level == "" {
		level = "INFO"
	}
	return &Log{level: level, prefix: prefix}
}

func (lr *Log) Report(n lattice.Node) error {
	lr.count++
	if lr.prefix != "" {
		errors.Logf(lr.level, "%s: sample %v %v", lr.prefix, lr.count, n)
	} else {
		errors.Logf(lr.level, "sample %v %v", lr.count, n)
	}
	return nil
}

func (lr *Log) Close() error {
	return nil
}
