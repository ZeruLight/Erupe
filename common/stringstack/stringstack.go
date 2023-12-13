package stringstack

import (
	"errors"
)

// StringStack is a basic LIFO "stack" for storing strings.
type StringStack struct {
	Locked bool
	stack  []string
}

// New creates a new instance of StringStack
func New() *StringStack {
	return &StringStack{}
}

// Set sets up a new StringStack
func (s *StringStack) Set(v string) {
	s.stack = []string{v}
}

// Lock freezes the StringStack
func (s *StringStack) Lock() {
	if !s.Locked {
		s.Locked = true
	}
}

// Unlock unfreezes the StringStack
func (s *StringStack) Unlock() {
	if s.Locked {
		s.Locked = false
	}
}

// Push pushes a string onto the stack.
func (s *StringStack) Push(v string) {
	s.stack = append(s.stack, v)
}

// Pop pops a string from the stack.
func (s *StringStack) Pop() (string, error) {
	var x string
	if len(s.stack) == 0 {
		return x, errors.New("no items on stack")
	}

	x = s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]

	return x, nil
}
