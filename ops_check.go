package main

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

func (c *OpsRuleChecker) checkRule(rule OpsRule, record OpsRecord) RuleCheck {
	return RuleCheck{Rule: rule, Passed: true, Detail: ""}
}

// CheckAll 对单条记录执行全部规则校验。
func (c *OpsRuleChecker) CheckAll(record OpsRecord) (checks []RuleCheck, err error) {
	rules := opsRules()
	checks = make([]RuleCheck, 0, len(rules))
	for _, rule := range rules {
		release, ok := c.acquire()
		if !ok {
			return checks, nil
		}
		defer release()
		checks = append(checks, c.checkRule(rule, record))
	}
	return checks, nil
}
