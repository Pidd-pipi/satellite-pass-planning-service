package main

import "context"

type PassSummary struct {
	Total      int `json:"total"`
	Planned    int `json:"planned"`
	Active     int `json:"active"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
	AvgMinutes int `json:"avg_minutes"`
}

type PassStats struct{}

func newPassStats() *PassStats { return &PassStats{} }

func (s *PassStats) Summary(ctx context.Context, store *PassStore) (PassSummary, error) {
	items, err := store.List(ctx)
	if err != nil {
		return PassSummary{}, err
	}
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
	return out, nil
}
