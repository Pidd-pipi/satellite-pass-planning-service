package main

import "testing"

// 规则列表里的 Code 必须全局唯一。
func TestOpsRuleCodesUnique(t *testing.T) {
	seen := map[string]string{}
	for _, rule := range opsRules() {
		if rule.Code == "" {
			t.Fatal("rule with empty code")
		}
		if prev, ok := seen[rule.Code]; ok {
			t.Fatalf("duplicate rule code %s (%s vs %s)", rule.Code, prev, rule.Name)
		}
		seen[rule.Code] = rule.Name
	}
}

// 规则严重度必须落在 low/normal/high/critical。
func TestOpsRuleSeverityValid(t *testing.T) {
	for _, rule := range opsRules() {
		switch rule.Severity {
		case OpsPriorityLow, OpsPriorityNormal, OpsPriorityHigh, OpsPriorityCritical:
		default:
			t.Fatalf("rule %s has invalid severity %q", rule.Code, rule.Severity)
		}
	}
}

// 过站规划站点规则必须引用正确站点且时长上限正确。
func TestPassRuleForStationLimits(t *testing.T) {
	cases := []struct {
		station string
		max     int
	}{
		{"North Ridge", 180},
		{"West Mesa", 240},
		{"East Yard", 120},
	}
	for _, c := range cases {
		rule, ok := passRuleForStation(c.station)
		if !ok {
			t.Fatalf("passRuleForStation(%s) not found", c.station)
		}
		if rule.MaxMinutes != c.max {
			t.Fatalf("%s MaxMinutes = %d, want %d", c.station, rule.MaxMinutes, c.max)
		}
	}
}

// 过站规划规则表必须覆盖全部已知站点且编码格式正确。
func TestPassRuleStationsWellFormed(t *testing.T) {
	known := map[string]bool{"North Ridge": false, "West Mesa": false, "East Yard": false, "South Loop": false}
	for _, rule := range passRuleDefinitions {
		if _, ok := known[rule.AppliesTo]; ok {
			known[rule.AppliesTo] = true
		}
		if rule.MaxMinutes <= 0 {
			t.Fatalf("rule %s has non-positive MaxMinutes", rule.Code)
		}
	}
	for station, found := range known {
		if !found {
			t.Fatalf("pass rule table missing station %q (or misspelled)", station)
		}
	}
}
