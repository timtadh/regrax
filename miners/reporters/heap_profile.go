package reporters

import (
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

import ()

import (
	"github.com/timtadh/sfp/lattice"
)

type HeapProfile struct {
	after, every, count int
	f                   io.WriteCloser
}

func NewHeapProfile(path string, after, every int) (*HeapProfile, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	hp := &HeapProfile{after: after, every: every, f: f}
	return hp, nil
}

func (hp *HeapProfile) Report(n lattice.Node) error {
	hp.count++
	if hp.count > hp.after && hp.count%hp.every == 0 {
		runtime.GC()
		runtime.GC()
		runtime.GC()
		return pprof.WriteHeapProfile(hp.f)
	} else {
		return nil
	}
}

func (hp *HeapProfile) Close() error {
	return hp.f.Close()
}
