package digraph

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

type Collector struct {
	MaxVertices int
	done *sync.Cond
	processed int
	collection SubGraphs
	requests chan *goiso.SubGraph
}

func NewCollector(maxVertices int) *Collector {
	c := &Collector{
		MaxVertices: maxVertices,
		done: sync.NewCond(new(sync.Mutex)),
		collection: make(SubGraphs, 0, 10),
		requests: make(chan *goiso.SubGraph),
	}
	go c.work()
	return c
}

func (c *Collector) work() {
	for sg := range c.requests {
		c.done.L.Lock()
		if len(sg.V) <= c.MaxVertices {
			c.collection = append(c.collection, sg)
		}
		c.processed++
		c.done.L.Unlock()
		c.done.Signal()
	}
}

func (c *Collector) Ch() chan *goiso.SubGraph {
	return c.requests
}

func (c *Collector) Wait(till int) []SubGraphs {
	c.done.L.Lock()
	defer c.done.L.Unlock()
	for c.processed < till {
		c.done.Wait()
	}
	close(c.requests)
	return c.collection.Partition()
}

