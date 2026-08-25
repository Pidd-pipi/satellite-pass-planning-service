package main

import (
	"fmt"
	"sync"
)

var passTransitionTable = map[string]map[string]bool{
	"planned":   {"active": true, "cancelled": true},
	"active":    {"completed": true, "cancelled": true},
	"completed": {},
	"cancelled": {},
}

type PassTransition struct {
	From   string
	To     string
	Reason string
}

type PassStateMachine struct {
	mu      sync.RWMutex
	history []PassTransition
}

func newPassStateMachine() *PassStateMachine { return &PassStateMachine{history: []PassTransition{}} }

func (m *PassStateMachine) CanMove(from, to string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return from == to || passTransitionTable[from][to]
}

func (m *PassStateMachine) Move(from, to, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if from == to {
		return nil
	}
	if !passTransitionTable[from][to] {
		return fmt.Errorf("%w: pass %s to %s", ErrPassTransition, from, to)
	}
	m.history = append(m.history, PassTransition{From: from, To: to, Reason: reason})
	return nil
}

func (m *PassStateMachine) History() []PassTransition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]PassTransition(nil), m.history...)
}

func (m *PassStateMachine) Last() (PassTransition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.history) == 0 {
		return PassTransition{}, false
	}
	return m.history[len(m.history)-1], true
}

func (m *PassStateMachine) Reset() { m.mu.Lock(); defer m.mu.Unlock(); m.history = m.history[:0] }

func passStateValid(value string) bool {
	switch value {
	case "planned", "active", "completed", "cancelled":
		return true
	}
	return false
}

func passStateTerminal(value string) bool { return value == "completed" || value == "cancelled" }
