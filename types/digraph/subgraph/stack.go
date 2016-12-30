package subgraph

import (
	"sync"
)

import (
)

import ()

type embSearchNode struct {
	ids *IdNode
	eid int
}

type Stack struct {
	mu sync.RWMutex
	cond  *sync.Cond
	stack []embSearchNode
	threads int
	waiting int
	prune func(*IdNode) bool
	closed bool
}

func NewStack(threads int, prune func(*IdNode) bool) *Stack {
	s := &Stack{
		stack: make([]embSearchNode, 0, 10),
		threads: threads,
		prune: prune,
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Stack) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.stack) == 0
}

func (s *Stack) Close() {
	s.mu.Lock()
	s.closed = true
	s.stack = nil
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Stack) Push(ids *IdNode, eid int) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	if s.prune != nil && s.prune(ids) {
		s.mu.Unlock()
		return
	}
	s.stack = append(s.stack, embSearchNode{ids, eid})
	s.mu.Unlock()
	s.cond.Signal()
}

func (s *Stack) Pop() (ids *IdNode, eid int) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, 0
	}
	s.waiting++
	for len(s.stack) == 0 {
		if s.waiting == s.threads || s.closed {
			s.mu.Unlock()
			s.cond.Broadcast()
			return nil, 0
		}
		s.cond.Wait()
	}
	s.waiting--
	item := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	s.mu.Unlock()
	return item.ids, item.eid
}
