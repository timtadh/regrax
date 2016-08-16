package vsigram

import (
	"sync"
)

import (
)

import (
	"github.com/timtadh/sfp/lattice"
)

type Stack struct {
	mu sync.RWMutex
	stack []lattice.Node
}

func NewStack() *Stack {
	return &Stack{
		stack: make([]lattice.Node, 0, 10),
	}
}

func (s *Stack) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.stack) == 0
}

func (s *Stack) Push(item lattice.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stack = append(s.stack, item)
}

func (s *Stack) Pop() (item lattice.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.stack) == 0 {
		return nil
	}
	item = s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return item
}
