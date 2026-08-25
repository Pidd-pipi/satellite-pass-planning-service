package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// opsContext 必须传播父上下文的取消。
func TestOpsContextPropagatesParentCancel(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	derived, derivedCancel := opsContext(parent, 5*time.Second)
	defer derivedCancel()
	cancel()
	if derived.Err() == nil {
		t.Fatal("opsContext dropped parent cancellation")
	}
}

// opsDelay 在上下文取消时必须立即返回。
func TestOpsDelayHonorsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- opsDelay(ctx, 2*time.Second) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("opsDelay returned nil after context cancel")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("opsDelay blocked after context cancel")
	}
}

// transition 入口必须使用请求自身的上下文。
func TestOpsTransitionContextUsesRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/ops/records/ops-001/transition", nil)
	req = req.WithContext(ctx)
	if got := opsTransitionContext(req); got.Err() == nil {
		t.Fatal("opsTransitionContext must return the request context, not Background")
	}
}

// 客户端在 transition 执行中取消请求，状态不得被推进且不写审计。
func TestOpsTransitionCanceledMidFlight(t *testing.T) {
	server := httptest.NewServer(NewHandler(NewPassStore()))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	after := time.AfterFunc(50*time.Millisecond, cancel)
	defer after.Stop()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		server.URL+"/ops/records/ops-001/transition",
		bytes.NewBufferString(`{"target":"active","expected":1}`))
	if err != nil {
		t.Fatal(err)
	}
	resp, doErr := http.DefaultClient.Do(req)
	if doErr == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal("transition completed after client cancellation")
		}
	}

	time.Sleep(600 * time.Millisecond)
	auditReq, _ := http.NewRequest(http.MethodGet, server.URL+"/ops/records/ops-001/audit", nil)
	auditResp, err := http.DefaultClient.Do(auditReq)
	if err != nil {
		t.Fatal(err)
	}
	defer auditResp.Body.Close()
	var payload struct {
		Events []OpsEvent `json:"events"`
	}
	if err := json.NewDecoder(auditResp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	for _, event := range payload.Events {
		if event.Type == "status_changed" {
			t.Fatalf("status_changed audit recorded despite client cancellation")
		}
	}
}
