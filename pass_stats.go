package main

import (
	"context"
	"sync"
)

type PassSummary struct {
	Total      int `json:"total"`
	Planned    int `json:"planned"`
	Active     int `json:"active"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
	AvgMinutes int `json:"avg_minutes"`
}

type PassStats struct {
	mu     sync.Mutex
	cached *PassSummary
}

func newPassStats() *PassStats { return &PassStats{} }

func computePassSummary(items []PassWindow) PassSummary {
	out := PassSummary{}
	totalMinutes := 0
	for _, item := range items {
		out.Total++
		switch item.State {
		case "planned":
			out.Planned++
		case "active":
			out.Active++
		case "completed":
			out.Completed++
		case "cancelled":
			out.Cancelled++
		}
		minutes := int(item.End.Sub(item.Start).Minutes())
		if minutes > 0 {
			totalMinutes += minutes
		}
	}
	if out.Total > 0 {
		out.AvgMinutes = totalMinutes / out.Total
	}
	return out
}

func (s *PassStats) Summary(ctx context.Context, store *PassStore) (PassSummary, error) {
	items, err := store.List(ctx)
	if err != nil {
		return PassSummary{}, err
	}
	out := computePassSummary(items)
	s.mu.Lock()
	s.cached = &out
	s.mu.Unlock()
	return out, nil
}

func (s *PassStats) Cached() PassSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached == nil {
		return PassSummary{}
	}
	return *s.cached
}

// Invalidate drops the cached summary so subsequent readers do not observe a
// stale snapshot after the underlying passes change. The next Summary call
// (from the background ticker or an HTTP request) repopulates it.
func (s *PassStats) Invalidate() {
	s.mu.Lock()
	s.cached = nil
	s.mu.Unlock()
}
