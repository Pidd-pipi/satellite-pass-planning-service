package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// review 必须是合法状态。
func TestOpsStatusReviewValid(t *testing.T) {
	if !opsStatusValid(OpsStatusReview) {
		t.Fatal("OpsStatusReview must be a valid status")
	}
}

// review 不是终态。
func TestOpsReviewNotTerminal(t *testing.T) {
	if opsStatusTerminal(OpsStatusReview) {
		t.Fatal("OpsStatusReview must not be terminal")
	}
}

// active -> review 迁移必须被允许。
func TestOpsTransitionToReviewAllowed(t *testing.T) {
	machine := newOpsStateMachine()
	if err := machine.Move(OpsStatusActive, OpsStatusReview, "start review"); err != nil {
		t.Fatalf("active -> review rejected: %v", err)
	}
}

// review -> active 迁移必须被允许（复核退回）。
func TestOpsTransitionReviewBackToActive(t *testing.T) {
	machine := newOpsStateMachine()
	if err := machine.Move(OpsStatusReview, OpsStatusActive, "send back"); err != nil {
		t.Fatalf("review -> active rejected: %v", err)
	}
}

// review 属于进行中状态集合。
func TestOpsInProgressIncludesReview(t *testing.T) {
	if !opsInProgress(OpsStatusReview) {
		t.Fatal("review must count as in-progress")
	}
}

// 通过公开接口把 active 记录流转到 review 必须成功。
func TestOpsTransitionReviewViaHTTP(t *testing.T) {
	handler := NewHandler(NewPassStore())
	req := httptest.NewRequest(http.MethodPost, "/ops/records/ops-002/transition",
		bytes.NewBufferString(`{"target":"review","expected":1}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("transition to review -> %d, want 200: %s", rec.Code, rec.Body.String())
	}
}
