package digraph

import (
	"sync"
	"math/rand"
)

type extWorkers struct {
	workers []*extWorker
	wg      sync.WaitGroup
}

func newExtWorkers(n int) *extWorkers {
	wkrs := &extWorkers{
		workers: make([]*extWorker, 0, n),
	}
	for i := 0; i < n; i++ {
		w := &extWorker{
			in: make(chan func()),
			wg: &wkrs.wg,
		}
		go w.work()
		wkrs.workers = append(wkrs.workers, w)
	}
	return wkrs
}

func (w *extWorkers) Stop() {
	workers := w.workers
	w.workers = nil
	for _, wrkr := range workers {
		close(wrkr.in)
	}
	w.wg.Wait()
}

func (w *extWorkers) Do(f func()) {
	workers := w.workers
	offset := rand.Intn(len(workers))
	for i := 0; i < len(workers); i++ {
		j := (offset + i) % len(workers)
		wrkr := workers[j].in
		select {
		case wrkr<-f:
			return
		default:
		}
	}
	workers[offset].in<-f
}

type extWorker struct {
	in chan func()
	wg *sync.WaitGroup
}

func (w *extWorker) work() {
	w.wg.Add(1)
	for f := range w.in {
		f()
	}
	w.wg.Done()
}
