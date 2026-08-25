package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var ErrInvalidPass = ErrPassInvalid

type PassStore struct {
	mu     sync.RWMutex
	next   int
	passes []PassWindow
}

func NewPassStore() *PassStore {
	now := time.Now().UTC()
	return &PassStore{
		next: 2,
		passes: []PassWindow{{
			ID: "pass-001", Satellite: "Aurora-7", Station: "North Ridge",
			Start: now.Add(20 * time.Minute), End: now.Add(32 * time.Minute), State: "planned",
		}},
	}
}

func (s *PassStore) List(ctx context.Context) ([]PassWindow, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return s.passes, nil
}

func (s *PassStore) Get(ctx context.Context, id string) (PassWindow, error) {
	select {
	case <-ctx.Done():
		return PassWindow{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.passes {
		if s.passes[i].ID == id {
			return s.passes[i], nil
		}
	}
	return PassWindow{}, ErrPassNotFound
}

func (s *PassStore) Add(ctx context.Context, pass PassWindow) (PassWindow, error) {
	select {
	case <-ctx.Done():
		return PassWindow{}, ctx.Err()
	default:
	}
	if pass.Satellite == "" || pass.Station == "" || pass.End.Before(pass.Start) {
		return PassWindow{}, ErrInvalidPass
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	pass.ID = fmt.Sprintf("pass-%03d", s.next)
	s.next++
	s.passes = append(s.passes, pass)
	return pass, nil
}

func (s *PassStore) UpdateState(ctx context.Context, id, state string) (PassWindow, error) {
	select {
	case <-ctx.Done():
		return PassWindow{}, ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.passes {
		if s.passes[i].ID == id {
			s.passes[i].State = state
			return s.passes[i], nil
		}
	}
	return PassWindow{}, ErrPassNotFound
}
