package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func postJSON(handler http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// 对不存在的 ops 记录做 transition 必须返回 404。
func TestOpsTransitionMissingRecordReturns404(t *testing.T) {
	handler := NewHandler(NewPassStore())
	rec := postJSON(handler, "/ops/records/ops-999/transition", `{"target":"active","expected":1}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("transition missing record -> %d, want 404 (error chain broken -> 500): %s", rec.Code, rec.Body.String())
	}
}

// 版本不匹配的 transition 必须返回 409。
func TestOpsTransitionConflictReturns409(t *testing.T) {
	handler := NewHandler(NewPassStore())
	rec := postJSON(handler, "/ops/records/ops-002/transition", `{"target":"active","expected":99}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("stale revision transition -> %d, want 409", rec.Code)
	}
}

// 缺少必填 owner 的创建必须返回 403（策略拒绝）。
func TestOpsCreatePolicyForbidden(t *testing.T) {
	handler := NewHandler(NewPassStore())
	rec := postJSON(handler, "/ops/records", `{"id":"ops-777","subject":"no owner","status":"queued","priority":"normal","labels":{"site":"gs-1"}}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("policy violation -> %d, want 403", rec.Code)
	}
}

// 存储层错误必须可被 errors.Is 识别为 not found。
func TestOpsStoreGetErrorChain(t *testing.T) {
	store := newOpsStore(seedOpsRecords())
	_, err := store.Get(context.Background(), "ops-404")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("errors.Is(err, ErrOpsNotFound) = false for %v (chain broken)", err)
	}
}

// opsCode 必须能识别包装后的哨兵错误。
func TestOpsCodeNotFoundClassification(t *testing.T) {
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrOpsNotFound))
	if code := opsCode(wrapped); code != "not_found" {
		t.Fatalf("opsCode(wrapped not found) = %q, want not_found", code)
	}
}

// 存储层 Update 的版本冲突错误必须可被 errors.Is 识别。
func TestOpsStoreUpdateConflictChain(t *testing.T) {
	store := newOpsStore(seedOpsRecords())
	record, err := store.Get(context.Background(), "ops-001")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Update(context.Background(), record, 99)
	if err == nil {
		t.Fatal("expected revision conflict")
	}
	if !errors.Is(err, ErrOpsConflict) {
		t.Fatalf("errors.Is(err, ErrOpsConflict) = false for %v (chain broken)", err)
	}
}

// opsCode 必须能识别包装后的 conflict / invalid / transition 哨兵错误。
func TestOpsCodeConflictClassification(t *testing.T) {
	if code := opsCode(fmt.Errorf("outer: %w", ErrOpsConflict)); code != "conflict" {
		t.Fatalf("opsCode(wrapped conflict) = %q, want conflict", code)
	}
}

func TestOpsCodeInvalidClassification(t *testing.T) {
	if code := opsCode(fmt.Errorf("outer: %w", ErrOpsInvalid)); code != "invalid" {
		t.Fatalf("opsCode(wrapped invalid) = %q, want invalid", code)
	}
}

func TestOpsCodeTransitionClassification(t *testing.T) {
	if code := opsCode(fmt.Errorf("outer: %w", ErrOpsTransition)); code != "transition" {
		t.Fatalf("opsCode(wrapped transition) = %q, want transition", code)
	}
}
