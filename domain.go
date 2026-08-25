package main

import (
	"errors"
	"time"
)

var (
	ErrPassNotFound   = errors.New("pass window not found")
	ErrPassTransition = errors.New("pass status transition is not allowed")
	ErrPassConflict   = errors.New("pass window conflicts with an existing window")
	ErrPassInvalid    = errors.New("pass request is invalid")
	ErrPassPolicy     = errors.New("pass planning policy rejected the request")
)

type PassWindow struct {
	ID        string    `json:"id"`
	Satellite string    `json:"satellite"`
	Station   string    `json:"station"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	State     string    `json:"state"`
}

func (p PassWindow) Clone() PassWindow { return p }

type CreatePassRequest struct {
	Satellite string `json:"satellite"`
	Station   string `json:"station"`
	Minutes   int    `json:"minutes"`
}

type CancelPassRequest struct {
	Reason string `json:"reason"`
	Actor  string `json:"actor"`
}

type ExportPassRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ExportPassResult struct {
	Exported int      `json:"exported"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	IDs      []string `json:"ids"`
}
