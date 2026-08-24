package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type passAPI struct {
	store    *PassStore
	audit    *PassAudit
	state    *PassStateMachine
	planner  *PassPlanner
	stats    *PassStats
	exporter *PassExporter
	notifier *PassNotifier
	reporter *PassReporter
}

func newPassAPI(store *PassStore) *passAPI {
	audit := newPassAudit()
	return &passAPI{
		store:    store,
		audit:    audit,
		state:    newPassStateMachine(),
		planner:  newPassPlanner(newOpsClock(), LoadConfig().Limits),
		stats:    newPassStats(),
		exporter: newPassExporter(store, audit),
		notifier: newPassNotifier(),
		reporter: newPassReporter(),
	}
}

func passStatusForError(err error) int {
	switch {
	case errors.Is(err, ErrPassNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrPassTransition):
		return http.StatusConflict
	case errors.Is(err, ErrPassConflict):
		return http.StatusConflict
	case errors.Is(err, ErrPassInvalid):
		return http.StatusBadRequest
	case errors.Is(err, ErrPassPolicy):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (a *passAPI) list(w http.ResponseWriter, r *http.Request) {
	q := PassQuery{
		Satellite: r.URL.Query().Get("satellite"),
		Station:   r.URL.Query().Get("station"),
		State:     r.URL.Query().Get("state"),
	}
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		q.Page = page
	}
	if size, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil {
		q.PageSize = size
	}
	items, err := a.store.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	filtered := passFilterBySatellite(items, q.Satellite)
	filtered = passFilterByState(items, q.State)
	sortPasses(filtered)
	q = passQueryDefaults(q)
	start, end := passBounds(len(filtered), q.Page, q.PageSize)
	writeJSON(w, http.StatusOK, PassPage{Items: filtered[start:end], Page: q.Page, PageSize: q.PageSize, Total: len(filtered), HasNext: end < len(filtered)})
}

func (a *passAPI) create(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCreate(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := a.planner.EnforceStationLimit(r.Context(), a.store, req.Station); err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	pass, err := a.planner.Plan(r.Context(), req)
	if err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	created, err := a.store.Add(r.Context(), pass)
	if err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	event := a.audit.Add(created.ID, "created", opsActorFromRequest(r))
	a.notifier.Publish(event)
	writeJSON(w, http.StatusCreated, created)
}

func (a *passAPI) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := a.store.Get(r.Context(), id)
	if err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *passAPI) cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := a.store.Get(r.Context(), id)
	if err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	if err := a.state.Move(item.State, "cancelled", "operator cancel"); err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	updated, err := a.store.UpdateState(r.Context(), id, "cancelled")
	if err != nil {
		writeJSON(w, passStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	event := a.audit.Add(id, "cancelled", opsActorFromRequest(r))
	a.notifier.Publish(event)
	writeJSON(w, http.StatusOK, updated)
}

func (a *passAPI) auditEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]any{"events": a.audit.For(id)})
}

func (a *passAPI) summary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.stats.Summary(r.Context(), a.store)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (a *passAPI) export(w http.ResponseWriter, r *http.Request) {
	var req ExportPassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	result, err := a.exporter.Export(ctx, req.From, req.To)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *passAPI) report(w http.ResponseWriter, r *http.Request) {
	report, err := a.reporter.Generate(r.Context(), a.store)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *passAPI) subscribe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ch := a.notifier.Subscribe(id)
	defer a.notifier.Unsubscribe(id, ch)
	select {
	case event := <-ch:
		writeJSON(w, http.StatusOK, event)
	case <-r.Context().Done():
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "subscription timed out"})
	}
}

func (a *passAPI) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/passes", a.list)
	mux.HandleFunc("POST /api/passes", a.create)
	mux.HandleFunc("GET /api/passes/summary", a.summary)
	mux.HandleFunc("POST /api/passes/export", a.export)
	mux.HandleFunc("GET /api/passes/report", a.report)
	mux.HandleFunc("GET /api/passes/{id}", a.get)
	mux.HandleFunc("POST /api/passes/{id}/cancel", a.cancel)
	mux.HandleFunc("GET /api/passes/{id}/audit", a.auditEvents)
	mux.HandleFunc("GET /api/passes/{id}/subscribe", a.subscribe)
	return mux
}
