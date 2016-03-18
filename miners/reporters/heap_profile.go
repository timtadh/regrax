package reporters

import (
	"os"
	"io"
	"runtime/pprof"
)

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)

type HeapProfile struct {
	f io.WriteCloser
}

func NewHeapProfile(path string) (*HeapProfile, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	hp := &HeapProfile{f: f}
	return hp, nil
}

func (hp *HeapProfile) Report(n lattice.Node) error {
	return pprof.WriteHeapProfile(hp.f)
}

func (hp *HeapProfile) Close() error {
	return hp.f.Close()
}

