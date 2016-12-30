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
	stack []embSearchNode
}

func NewStack() *Stack {
	return &Stack{
		stack: make([]embSearchNode, 0, 10),
	}
}

func (s *Stack) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.stack) == 0
}

func (s *Stack) Push(ids *IdNode, eid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stack = append(s.stack, embSearchNode{ids, eid})
}

func (s *Stack) Pop() (ids *IdNode, eid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.stack) == 0 {
		return nil, 0
	}
	item := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return item.ids, item.eid
}
