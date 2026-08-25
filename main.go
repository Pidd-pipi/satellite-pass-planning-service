package main

import "log"

func main() {
	cfg := LoadConfig()
	store := NewPassStore()
	handler := NewHandler(store)
	defer stopStatsRefresher(handler)
	log.Printf("satellite pass planning service listening on :%s", cfg.Port)
	if err := serveAddress(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}

// stopStatsRefresher stops the background stats ticker when the handler exposes
// a Close method, preventing the goroutine and its timer from outliving the
// server. log.Fatal bypasses deferred calls, but the process is exiting then
// regardless; the normal shutdown path (SIGINT/SIGTERM) returns here.
func stopStatsRefresher(handler any) {
	if c, ok := handler.(interface{ Close() error }); ok {
		_ = c.Close()
	}
}
