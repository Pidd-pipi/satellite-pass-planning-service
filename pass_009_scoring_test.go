package main

import (
	"context"
	"testing"
	"time"
)

// 状态机历史必须被限制，不能无限增长。
func TestPassStateHistoryBounded(t *testing.T) {
	machine := newPassStateMachine()
	for i := 0; i < 150; i++ {
		if err := machine.Move("planned", "active", "start"); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(machine.History()); n > 100 {
		t.Fatalf("state history grew to %d, want capped at %d", n, 100)
	}
}

// 状态机历史返回必须隔离，外部改写不得污染内部。
func TestPassStateHistoryIsolated(t *testing.T) {
	machine := newPassStateMachine()
	if err := machine.Move("planned", "active", "start"); err != nil {
		t.Fatal(err)
	}
	history := machine.History()
	if len(history) == 0 {
		t.Fatal("expected history entry")
	}
	history[0].Reason = "tampered"
	if again := machine.History(); again[0].Reason == "tampered" {
		t.Fatal("mutating returned history polluted the state machine")
	}
}

// 取消订阅必须关闭通知通道，否则订阅 goroutine 泄漏。
func TestPassNotifyUnsubscribeClosesChannel(t *testing.T) {
	notifier := newPassNotifier()
	ch := notifier.Subscribe("pass-001")
	notifier.Publish(PassAuditEvent{PassID: "pass-001", Type: "created"})
	notifier.Unsubscribe("pass-001", ch)
	closed := make(chan struct{})
	go func() {
		for {
			if _, ok := <-ch; !ok {
				close(closed)
				return
			}
		}
	}()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("subscriber channel never closed after unsubscribe (goroutine leak)")
	}
}

// 报告历史必须被限制，不能无限累积。
func TestPassReportHistoryBounded(t *testing.T) {
	reporter := newPassReporter()
	store := NewPassStore()
	ctx := context.Background()
	for i := 0; i < 150; i++ {
		if _, err := reporter.Generate(ctx, store); err != nil {
			t.Fatal(err)
		}
	}
	if n := reporter.Count(); n > 100 {
		t.Fatalf("report history grew to %d, want capped at %d", n, 100)
	}
}
