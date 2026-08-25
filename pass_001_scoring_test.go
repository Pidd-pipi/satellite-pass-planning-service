package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func seedPasses(t *testing.T, handler http.Handler, bodies ...string) {
	t.Helper()
	for _, body := range bodies {
		req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %s -> %d: %s", body, rec.Code, rec.Body.String())
		}
	}
}

func getPassPage(t *testing.T, handler http.Handler, url string) PassPage {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s -> %d", url, rec.Code)
	}
	var page PassPage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	return page
}

// 组合过滤必须只返回同时命中卫星和状态的窗口，不能混入不匹配项或重复项。
func TestPassListCombinedFilterCorrect(t *testing.T) {
	handler := NewHandler(NewPassStore())
	seedPasses(t, handler,
		`{"satellite":"Alpha-1","station":"West Mesa","minutes":10}`,
		`{"satellite":"Beta-2","station":"North Ridge","minutes":12}`,
		`{"satellite":"Alpha-1","station":"East Yard","minutes":15}`,
		`{"satellite":"Gamma-3","station":"South Loop","minutes":20}`,
	)
	page := getPassPage(t, handler, "/api/passes?satellite=Alpha-1&state=planned")
	if page.Total != 2 {
		t.Fatalf("combined filter total = %d, want 2", page.Total)
	}
	seen := map[string]bool{}
	for _, item := range page.Items {
		if item.Satellite != "Alpha-1" || item.State != "planned" {
			t.Fatalf("combined filter returned non-matching window: %+v", item)
		}
		if seen[item.ID] {
			t.Fatalf("combined filter returned duplicate window: %s", item.ID)
		}
		seen[item.ID] = true
	}
}

// 过滤函数不得改写输入切片的底层数组。
func TestPassFilterInputNotMutated(t *testing.T) {
	items := []PassWindow{
		{ID: "p1", Satellite: "Alpha-1", Station: "West Mesa", State: "planned"},
		{ID: "p2", Satellite: "Beta-2", Station: "North Ridge", State: "planned"},
		{ID: "p3", Satellite: "Alpha-1", Station: "East Yard", State: "active"},
	}
	original := make([]PassWindow, len(items))
	copy(original, items)
	_ = passFilterBySatellite(items, "Alpha-1")
	for i := range original {
		if items[i].ID != original[i].ID {
			t.Fatalf("passFilterByStation mutated input at index %d: %s -> %s", i, original[i].ID, items[i].ID)
		}
	}
	// 状态过滤同样不得回写输入：把不匹配项放在最前，回写必然破坏原数组。
	stateItems := []PassWindow{
		{ID: "q1", Satellite: "Alpha-1", Station: "West Mesa", State: "active"},
		{ID: "q2", Satellite: "Beta-2", Station: "North Ridge", State: "planned"},
		{ID: "q3", Satellite: "Alpha-1", Station: "East Yard", State: "planned"},
	}
	stateOriginal := make([]PassWindow, len(stateItems))
	copy(stateOriginal, stateItems)
	_ = passFilterByState(stateItems, "planned")
	for i := range stateOriginal {
		if stateItems[i].ID != stateOriginal[i].ID {
			t.Fatalf("passFilterByState mutated input at index %d: %s -> %s", i, stateOriginal[i].ID, stateItems[i].ID)
		}
	}
}

// 存储层 List 必须返回独立副本。
func TestPassStoreListReturnsCopy(t *testing.T) {
	store := NewPassStore()
	ctx := context.Background()
	first, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Fatal("expected seed")
	}
	first[0].State = "cancelled"
	second, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range second {
		if item.State == "cancelled" {
			t.Fatalf("store.List returned shared backing array: window %s became cancelled", item.ID)
		}
	}
}
