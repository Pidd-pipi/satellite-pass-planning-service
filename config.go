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
	limits := map[string]int{
		"max_window_minutes":      240,
		"max_planned_per_station": 3,
	}
	if v := os.Getenv("MAX_WINDOW_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits["max_window_minutes"] = n
		}
	}
	return Config{Port: port, Limits: limits}
}
