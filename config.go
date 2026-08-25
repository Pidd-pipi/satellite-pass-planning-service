package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port   string
	Limits map[string]int
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port, Limits: loadLimits()}
}

// loadLimits reads the tunable pass-planning limits from the environment.
// The map is always non-nil so callers can safely read from it; an unset
// or non-positive variable falls back to the documented default.
func loadLimits() map[string]int {
	limits := map[string]int{
		"max_window_minutes":      envInt("MAX_WINDOW_MINUTES", 240),
		"max_planned_per_station": envInt("MAX_PLANNED_PER_STATION", 3),
	}
	return limits
}

// envInt parses an integer environment variable. A missing, blank, or
// non-positive value yields fallback, which keeps the limit effective
// instead of silently disabling it.
func envInt(name string, fallback int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
