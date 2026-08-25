package main

import (
	"fmt"
	"strings"
)

type RuleCheck struct {
	Rule   OpsRule
	Passed bool
	Detail string
}

type OpsRuleChecker struct {
	tokens chan struct{}
}

func newOpsRuleChecker() *OpsRuleChecker {
	return &OpsRuleChecker{tokens: make(chan struct{}, 16)}
}

func (c *OpsRuleChecker) acquire() (release func(), ok bool) {
	select {
	case c.tokens <- struct{}{}:
		var once bool
		return func() {
			if !once {
				once = true
				<-c.tokens
			}
		}, true
	default:
		return nil, false
	}
}

// checkRule evaluates a single rule against a record. A rule passes when every
// RequiredLabel is present with a non-empty value. Rules flagged Terminal only
// apply to records that have reached a terminal (closed) status; they are
// skipped for non-terminal records rather than failing them.
func (c *OpsRuleChecker) checkRule(rule OpsRule, record OpsRecord) RuleCheck {
	if rule.Terminal && !record.Terminal() {
		return RuleCheck{Rule: rule, Passed: true, Detail: "skipped: terminal rule on non-terminal record"}
	}
	missing := make([]string, 0)
	for _, label := range rule.RequiredLabels {
		if strings.TrimSpace(record.LabelValue(label)) == "" {
			missing = append(missing, label)
		}
	}
	if len(missing) > 0 {
		return RuleCheck{Rule: rule, Passed: false, Detail: "missing required labels: " + strings.Join(missing, ", ")}
	}
	return RuleCheck{Rule: rule, Passed: true, Detail: ""}
}

// CheckAll 对单条记录执行全部规则校验。
func (c *OpsRuleChecker) CheckAll(record OpsRecord) (checks []RuleCheck, err error) {
	rules := opsRules()
	checks = make([]RuleCheck, 0, len(rules))
	for _, rule := range rules {
		// Acquire and release within each iteration. The lease token throttles
		// concurrent CheckAll callers; a single call must never hold more than
		// one token, otherwise the pool (cap 16) is exhausted partway through.
		release, ok := c.acquire()
		if !ok {
			return checks, fmt.Errorf("%w: rule lease pool exhausted after %d of %d rules", ErrOpsPolicy, len(checks), len(rules))
		}
		checks = append(checks, c.checkRule(rule, record))
		release()
	}
	return checks, nil
}
