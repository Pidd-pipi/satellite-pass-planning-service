package main

import (
	"fmt"
	"strings"
)

type PassRule struct {
	Code       string
	Name       string
	Severity   string
	AppliesTo  string
	MaxMinutes int
}

var passRuleDefinitions = []PassRule{
	{Code: "PASS-0101", Name: "north ridge antenna duty", Severity: "high", AppliesTo: "North Ridge", MaxMinutes: 180},
	{Code: "PASS-0102", Name: "west mesa uplink power", Severity: "critical", AppliesTo: "West Mesa", MaxMinutes: 240},
	{Code: "PASS-0103", Name: "east yard scheduling", Severity: "normal", AppliesTo: "East Yard", MaxMinutes: 120},
	{Code: "PASS-0104", Name: "south loop receive", Severity: "low", AppliesTo: "South Loop", MaxMinutes: 90},
}

func passRuleForStation(station string) (PassRule, bool) {
	for _, rule := range passRuleDefinitions {
		if strings.EqualFold(rule.AppliesTo, station) {
			return rule, true
		}
	}
	return PassRule{}, false
}

func passRulesAllowed(station string, minutes int) error {
	rule, ok := passRuleForStation(station)
	if !ok {
		return nil
	}
	if minutes > rule.MaxMinutes {
		return fmt.Errorf("%w: %s allows at most %d minutes", ErrPassPolicy, rule.Name, rule.MaxMinutes)
	}
	return nil
}

func passRuleCodes() []string {
	codes := make([]string, 0, len(passRuleDefinitions))
	for _, rule := range passRuleDefinitions {
		codes = append(codes, rule.Code)
	}
	return codes
}
