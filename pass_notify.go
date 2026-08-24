package main

import "sync"

type PassNotifier struct {
	mu          sync.Mutex
	subscribers map[string][]chan PassAuditEvent
}

func newPassNotifier() *PassNotifier {
	return &PassNotifier{subscribers: map[string][]chan PassAuditEvent{}}
}

func (n *PassNotifier) Subscribe(passID string) chan PassAuditEvent {
	ch := make(chan PassAuditEvent, 8)
	n.mu.Lock()
	defer n.mu.Unlock()
	n.subscribers[passID] = append(n.subscribers[passID], ch)
	return ch
}

func (n *PassNotifier) Unsubscribe(passID string, ch chan PassAuditEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs := n.subscribers[passID]
	for i, candidate := range subs {
		if candidate == ch {
			n.subscribers[passID] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

func (n *PassNotifier) Publish(event PassAuditEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, ch := range n.subscribers[event.PassID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (n *PassNotifier) SubscriberCount(passID string) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.subscribers[passID])
}

func (n *PassNotifier) TotalSubscribers() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	total := 0
	for _, subs := range n.subscribers {
		total += len(subs)
	}
	return total
}
