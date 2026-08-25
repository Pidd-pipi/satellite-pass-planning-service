package main

import (
	"context"
	"fmt"
	"time"
)

type PassPlanner struct {
	clock  OpsClock
	limits map[string]int
}

func newPassPlanner(clock OpsClock, limits map[string]int) *PassPlanner {
	return &PassPlanner{clock: clock, limits: limits}
}

func (p *PassPlanner) limitValue(key string, fallback int) int {
	if p.limits == nil {
		return fallback
	}
	if v, ok := p.limits[key]; ok && v > 0 {
		return v
	}
	return fallback
}

func (p *PassPlanner) maxWindowMinutes() int     { return p.limitValue("max_window_minutes", 240) }
func (p *PassPlanner) maxPlannedPerStation() int { return p.limitValue("max_planned_per_station", 3) }

func (p *PassPlanner) Plan(ctx context.Context, req CreatePassRequest) (PassWindow, error) {
	if req.Satellite == "" || req.Station == "" || req.Minutes <= 0 {
		return PassWindow{}, ErrInvalidPass
	}
	if err := passRulesAllowed(req.Station, req.Minutes); err != nil {
		return PassWindow{}, err
	}
	if req.Minutes > p.maxWindowMinutes() {
		return PassWindow{}, fmt.Errorf("%w: window longer than %d minutes", ErrPassInvalid, p.maxWindowMinutes())
	}
	start := p.clock.Now().Add(45 * time.Minute)
	return PassWindow{
		Satellite: req.Satellite,
		Station:   req.Station,
		Start:     start,
		End:       start.Add(time.Duration(45+req.Minutes) * time.Minute),
		State:     "planned",
	}, nil
}

func (p *PassPlanner) EnforceStationLimit(ctx context.Context, store *PassStore, station string) error {
	items, err := store.List(ctx)
	if err != nil {
		return err
	}
	planned := 0
	for _, item := range items {
		if item.Station == station && item.State == "planned" {
			planned++
		}
	}
	if planned >= p.maxPlannedPerStation() {
		return fmt.Errorf("%w: station %s already has %d planned windows", ErrPassConflict, station, p.maxPlannedPerStation())
	}
	return nil
}
