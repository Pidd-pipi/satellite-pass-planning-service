package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func getOpsRules2(t *testing.T, handler http.Handler) []OpsRule {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/ops/rules", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ops/rules -> %d", rec.Code)
	}
	var payload struct {
		Rules []OpsRule `json:"rules"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Rules
}

// 规则编码必须符合 OPS-\\d{4}、无空白且全局唯一。
func TestOpsRulesCodesCleanUnique(t *testing.T) {
	handler := NewHandler(NewPassStore())
	rules := getOpsRules2(t, handler)
	pattern := regexp.MustCompile(`^OPS-\d{4}$`)
	seen := map[string]bool{}
	for _, rule := range rules {
		if !pattern.MatchString(rule.Code) {
			t.Fatalf("rule code %q is malformed", rule.Code)
		}
		if seen[rule.Code] {
			t.Fatalf("duplicate rule code %q", rule.Code)
		}
		seen[rule.Code] = true
	}
}

// 所有规则严重度必须合法。
func TestOpsRulesSeveritiesValid(t *testing.T) {
	handler := NewHandler(NewPassStore())
	for _, rule := range getOpsRules2(t, handler) {
		switch rule.Severity {
		case OpsPriorityLow, OpsPriorityNormal, OpsPriorityHigh, OpsPriorityCritical:
		default:
			t.Fatalf("rule %s has invalid severity %q", rule.Code, rule.Severity)
		}
	}
}

// 规则名称不能为空。
func TestOpsRulesNamesNonEmpty(t *testing.T) {
	handler := NewHandler(NewPassStore())
	for _, rule := range getOpsRules2(t, handler) {
		if rule.Name == "" {
			t.Fatalf("rule %s has empty name", rule.Code)
		}
	}
}

// 每条规则必须声明必填标签（不能为 nil）。
func TestOpsRulesLabelsNonNil(t *testing.T) {
	handler := NewHandler(NewPassStore())
	for _, rule := range getOpsRules2(t, handler) {
		if rule.RequiredLabels == nil {
			t.Fatalf("rule %s has nil RequiredLabels", rule.Code)
		}
	}
}
