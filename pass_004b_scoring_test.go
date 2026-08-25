package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 默认配置（未设置 MAX_WINDOW_MINUTES）下创建窗口必须成功，不能 panic。
func TestPassCreateNoEnvNoPanic(t *testing.T) {
	t.Setenv("MAX_WINDOW_MINUTES", "")
	handler := NewHandler(NewPassStore())
	req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBufferString(`{"satellite":"Beacon-2","station":"West Mesa","minutes":10}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create with default config -> %d (nil limits panic): %s", rec.Code, rec.Body.String())
	}
}

// LoadConfig 必须解析 MAX_WINDOW_MINUTES 环境变量。
func TestLoadConfigParsesEnvLimit(t *testing.T) {
	t.Setenv("MAX_WINDOW_MINUTES", "10")
	cfg := LoadConfig()
	if cfg.Limits == nil {
		t.Fatal("LoadConfig returned nil Limits map")
	}
	if cfg.Limits["max_window_minutes"] != 10 {
		t.Fatalf("max_window_minutes = %d, want 10 (env ignored)", cfg.Limits["max_window_minutes"])
	}
}

// planner 对 nil 限额表读默认值不得 panic。
func TestPassPlannerNilLimitsSafe(t *testing.T) {
	planner := newPassPlanner(newOpsClock(), nil)
	if v := planner.maxWindowMinutes(); v != 240 {
		t.Fatalf("maxWindowMinutes = %d, want 240", v)
	}
}

// MAX_WINDOW_MINUTES=10 时，20 分钟的窗口必须被拒绝。
func TestPassCreateEnforcesEnvLimit(t *testing.T) {
	t.Setenv("MAX_WINDOW_MINUTES", "10")
	handler := NewHandler(NewPassStore())
	req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBufferString(`{"satellite":"Beacon-2","station":"West Mesa","minutes":20}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		// 被拒绝即为正确
		return
	}
	t.Fatal("20-minute window accepted when MAX_WINDOW_MINUTES=10")
}

// minutes 为 0 的请求必须被拒绝。
func TestPassCreateRejectsZeroMinutes(t *testing.T) {
	t.Setenv("MAX_WINDOW_MINUTES", "")
	handler := NewHandler(NewPassStore())
	req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBufferString(`{"satellite":"Beacon-2","station":"West Mesa","minutes":0}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("zero minutes -> %d, want 400", rec.Code)
	}
}

// North Ridge 站点规则 180 分钟上限必须生效。
func TestPassCreateRejectsStationRule(t *testing.T) {
	t.Setenv("MAX_WINDOW_MINUTES", "")
	handler := NewHandler(NewPassStore())
	req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBufferString(`{"satellite":"Beacon-2","station":"North Ridge","minutes":200}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("station rule violation -> %d, want 403", rec.Code)
	}
}
