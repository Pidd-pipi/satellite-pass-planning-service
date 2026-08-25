package main

import "testing"

func seedOpsRecord(labels map[string]string) OpsRecord {
	return OpsRecord{
		ID: "ops-check-1", Subject: "rule check subject", Owner: "li",
		Status: OpsStatusQueued, Priority: OpsPriorityNormal, Revision: 1,
		Labels: labels,
	}
}

// 全量 112 条规则必须完整跑完，不能因租约耗尽而失败。
func TestOpsRuleCheckAllComplete(t *testing.T) {
	checker := newOpsRuleChecker()
	record := seedOpsRecord(map[string]string{"site": "gs-1", "operator": "li", "evidence": "e1", "reviewed": "yes"})
	checks, err := checker.CheckAll(record)
	if err != nil {
		t.Fatalf("CheckAll failed: %v (lease pool exhausted by deferred releases)", err)
	}
	if len(checks) != 112 {
		t.Fatalf("checks = %d, want 112", len(checks))
	}
}

// 缺少必填 site 标签的记录必须被至少一条规则判为不通过。
func TestOpsRuleCheckDetectsMissingLabel(t *testing.T) {
	checker := newOpsRuleChecker()
	record := seedOpsRecord(map[string]string{"operator": "li", "evidence": "e1", "reviewed": "yes"})
	checks, err := checker.CheckAll(record)
	if err != nil {
		t.Fatal(err)
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		}
	}
	if passed == len(checks) {
		t.Fatal("missing site label was not detected by any rule (validation bypass)")
	}
}

// OPS-0101 的必填标签列表不能含拼写错误的标签。
func TestOpsRule0101LabelsCorrect(t *testing.T) {
	rule := opsRule0101()
	for _, label := range rule.RequiredLabels {
		if label == "sity" {
			t.Fatal("OPS-0101 references a misspelled label 'sity'")
		}
		switch label {
		case "site", "operator", "evidence", "reviewed":
		default:
			t.Fatalf("OPS-0101 has unexpected required label %q", label)
		}
	}
}

// OPS-0102 必须要求 site 标签，缺失时该规则必须判不通过。
func TestOpsRule0102RequiresSite(t *testing.T) {
	rule := opsRule0102()
	found := false
	for _, label := range rule.RequiredLabels {
		if label == "site" {
			found = true
		}
	}
	if !found {
		t.Fatal("OPS-0102 must require site label")
	}
}
