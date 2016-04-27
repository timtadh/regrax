package ext

import (
	"sync"
)

import (
	"github.com/timtadh/goiso"
)

type Collector struct {
	MaxVertices int
	done        *sync.Cond
	stopped     bool
	processed   int
	collection  Embeddings
	requests    chan *goiso.SubGraph
}

func NewCollector(maxVertices int) *Collector {
	c := &Collector{
		MaxVertices: maxVertices,
		done:        sync.NewCond(new(sync.Mutex)),
		collection:  make(Embeddings, 0, 10),
		requests:    make(chan *goiso.SubGraph),
	}
	go c.work()
	return c
}

func (c *Collector) work() {
	for sg := range c.requests {
		c.done.L.Lock()
		if len(sg.V) <= c.MaxVertices || c.MaxVertices < 0 {
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

func (c *Collector) Wait(till int) Embeddings {
	c.done.L.Lock()
	defer c.done.L.Unlock()
	for c.processed < till {
		c.done.Wait()
	}
	close(c.requests)
	c.stopped = true
	return c.collection
}
