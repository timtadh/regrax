package ext

import (
	"sync"
)

import (
	"github.com/timtadh/goiso"
)


type Extender struct {
	lock sync.Mutex
	stopped bool
	requests chan extRequest
}

type extRequest struct {
	sg *goiso.SubGraph
	e *goiso.Edge
	resp chan *goiso.SubGraph
}

func NewExtender(workers int) *Extender {
	x := &Extender{
		requests: make(chan extRequest),
	}
	for i := 0; i < workers; i++ {
		go x.work()
	}
	return x
}

func (x *Extender) Stop() {
	x.lock.Lock()
	defer x.lock.Unlock()
	if !x.stopped {
		close(x.requests)
		x.requests = nil
		x.stopped = true
	}
}

func (x *Extender) Extend(sg *goiso.SubGraph, e *goiso.Edge, resp chan *goiso.SubGraph) {
	x.requests<-extRequest{sg, e, resp}
}

func (x *Extender) work() {
	for req := range x.requests {
		req.resp<-x.extend(req.sg, req.e)
	}
}

func (x *Extender) extend(sg *goiso.SubGraph, e *goiso.Edge) *goiso.SubGraph {
	nsg, _ := sg.EdgeExtend(e)
	return nsg
}
