package main

import "net/http"

func NewHandler(store *PassStore) http.Handler {
	pass := newPassAPI(store)
	ops := newOpsAPIHandler(newOpsService(seedOpsRecords()))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.Handle("/api/passes", pass.routes())
	mux.Handle("/api/passes/", pass.routes())
	mux.Handle("/ops/", ops)
	mux.Handle("/ops", ops)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return &shutdownHandler{ServeMux: mux, pass: pass}
}

// shutdownHandler wraps the mux so the background stats ticker can be stopped
// cleanly when the server shuts down, preventing the goroutine and its timer
// from leaking.
type shutdownHandler struct {
	*http.ServeMux
	pass *passAPI
}

func (h *shutdownHandler) Close() error {
	if h.pass != nil {
		h.pass.Close()
	}
	return nil
}
