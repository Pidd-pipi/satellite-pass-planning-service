package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPassRoutes(t *testing.T) {
	server := httptest.NewServer(NewHandler(NewPassStore()))
	defer server.Close()
	resp, err := http.Get(server.URL + "/api/passes")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	resp.Body.Close()
	body := `{"satellite":"Beacon-2","station":"West Mesa","minutes":9}`
	resp, err = http.Post(server.URL+"/api/passes", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestPassRejectsInvalidPayload(t *testing.T) {
	server := httptest.NewServer(NewHandler(NewPassStore()))
	defer server.Close()
	resp, err := http.Post(server.URL+"/api/passes", "application/json", strings.NewReader(`{"station":"West Mesa","minutes":0}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	resp.Body.Close()
}
