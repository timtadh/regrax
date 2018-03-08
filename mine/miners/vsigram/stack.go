package vsigram

import (
	"sync"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/regrax/lattice"
)



type Stack struct {
	mu sync.Mutex
	cond  *sync.Cond
	stack []lattice.Node
	threads int
	waiting int
	closed bool
}

func NewStack(expectedThreads int) *Stack {
	s := &Stack{
		stack: make([]lattice.Node, 0, 100),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Stack) AddThread() int {
	s.mu.Lock()
	tid := s.threads
	s.threads++
	s.mu.Unlock()
	return tid
}

func (s *Stack) Close() {
	s.mu.Lock()
	s.closed = true
	s.stack = nil
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Stack) Closed() bool {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	return closed
}

func (s *Stack) WaitClosed() {
	s.mu.Lock()
	for !s.closed {
		s.cond.Wait()
	}
	s.mu.Unlock()
}

func (s *Stack) Push(tid int, node lattice.Node) {
	if false {
		errors.Logf("DEBUG", "tid %v", tid)
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.stack = append(s.stack, node)
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Stack) Pop(tid int) (node lattice.Node) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	for {
		if len(s.stack) > 0 {
			node = s.stack[len(s.stack)-1]
			s.stack = s.stack[:len(s.stack)-1]
			s.mu.Unlock()
			return node
		}

		// steal failed; wait for a broadcast of a Push
		s.waiting++
		if (s.threads > 0 && s.threads == s.waiting) || s.closed {
			s.mu.Unlock()
			s.Close()
			s.cond.Broadcast()
			return nil
		}
		s.cond.Wait()
		s.waiting--
	}
}
