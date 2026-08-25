package main

import (
	"context"
	"strings"
	"sync"
	"time"
)

type PassReport struct {
	GeneratedAt string
	Lines       []string
	Total       int
}

type PassReporter struct {
	mu      sync.Mutex
	history []PassReport
}

func newPassReporter() *PassReporter { return &PassReporter{} }

func (r *PassReporter) Generate(ctx context.Context, store *PassStore) (PassReport, error) {
	items, err := store.List(ctx)
	if err != nil {
		return PassReport{}, err
	}
	report := PassReport{GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano), Lines: []string{}, Total: len(items)}
	for _, item := range items {
		report.Lines = append(report.Lines, strings.Join([]string{
			item.ID, item.Satellite, item.Station, item.State,
		}, " | "))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history = append(r.history, report)
	return report, nil
}

func (r *PassReporter) History() []PassReport {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]PassReport, len(r.history))
	copy(out, r.history)
	return out
}

func (r *PassReporter) Count() int { r.mu.Lock(); defer r.mu.Unlock(); return len(r.history) }
