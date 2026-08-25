package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var passAuditSequence uint64

func newPassAuditID() string {
	return fmt.Sprintf("paevt-%06d", atomic.AddUint64(&passAuditSequence, 1))
}

type PassAuditEvent struct {
	ID      string            `json:"id"`
	PassID  string            `json:"pass_id"`
	Type    string            `json:"type"`
	Actor   string            `json:"actor"`
	At      string            `json:"at"`
	Details map[string]string `json:"details,omitempty"`
}

type PassAudit struct {
	mu     sync.RWMutex
	events []PassAuditEvent
}

func newPassAudit() *PassAudit { return &PassAudit{events: []PassAuditEvent{}} }

func (a *PassAudit) Add(passID, typ, actor string) PassAuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	event := PassAuditEvent{
		ID: newPassAuditID(), PassID: passID, Type: typ, Actor: actor,
		At: time.Now().UTC().Format(time.RFC3339Nano),
	}
	a.events = append(a.events, event)
	return event
}

func (a *PassAudit) For(passID string) []PassAuditEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []PassAuditEvent{}
	for _, event := range a.events {
		if event.PassID == passID {
			out = append(out, event)
		}
	}
	return out
}

func (a *PassAudit) Since(start time.Time) []PassAuditEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []PassAuditEvent{}
	for _, event := range a.events {
		parsed, err := time.Parse(time.RFC3339Nano, event.At)
		if err == nil && !parsed.Before(start) {
			out = append(out, event)
		}
	}
	return out
}

func (a *PassAudit) Count() int { a.mu.RLock(); defer a.mu.RUnlock(); return len(a.events) }

func (a *PassAudit) Latest() (PassAuditEvent, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.events) == 0 {
		return PassAuditEvent{}, false
	}
	return a.events[len(a.events)-1], true
}

func (a *PassAudit) Trim(max int) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if max < 0 || len(a.events) <= max {
		return 0
	}
	removed := len(a.events) - max
	a.events = append([]PassAuditEvent(nil), a.events[len(a.events)-max:]...)
	return removed
}

func (a *PassAudit) Clear() { a.mu.Lock(); defer a.mu.Unlock(); a.events = a.events[:0] }
