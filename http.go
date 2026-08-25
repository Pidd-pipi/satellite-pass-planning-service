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
	return mux
}
