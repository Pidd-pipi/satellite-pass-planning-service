package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type OpsPolicy struct {
	RequireOwner  bool
	RequiredLabel string
	MaxActive     int
}
type OpsService struct {
	store  *OpsStore
	audit  *OpsAudit
	state  *OpsStateMachine
	policy OpsPolicy
	clock  OpsClock
}

func newOpsService(seed []OpsRecord) *OpsService {
	return &OpsService{store: newOpsStore(seed), audit: newOpsAudit(), state: newOpsStateMachine(), policy: OpsPolicy{RequireOwner: true, RequiredLabel: "site", MaxActive: 1000}, clock: newOpsClock()}
}
func (p OpsPolicy) Check(record OpsRecord) error {
	if p.RequireOwner && strings.TrimSpace(record.Owner) == "" {
		return fmt.Errorf("%w: owner required", ErrOpsPolicy)
	}
	if p.RequiredLabel != "" && record.LabelValue(p.RequiredLabel) == "" {
		return fmt.Errorf("%w: %s label required", ErrOpsPolicy, p.RequiredLabel)
	}
	if record.Priority == "" {
		return fmt.Errorf("%w: priority required", ErrOpsPolicy)
	}
	return nil
}
func (s *OpsService) Create(ctx context.Context, record OpsRecord) (OpsRecord, error) {
	record = normalizeOpsRecord(record)
	if err := s.policy.Check(record); err != nil {
		return OpsRecord{}, err
	}
	record.CreatedAt = s.clock.Stamp()
	record.UpdatedAt = record.CreatedAt
	if err := s.store.Put(ctx, record); err != nil {
		return OpsRecord{}, wrapOps("create", "store.put", err)
	}
	s.audit.Add(record.ID, "created", record.Owner)
	return record, nil
}
func (s *OpsService) Get(ctx context.Context, id string) (OpsRecord, error) {
	return s.store.Get(ctx, id)
}
func (s *OpsService) Search(ctx context.Context, q OpsQuery) (OpsPage, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return OpsPage{}, err
	}
	filtered := make([]OpsRecord, 0, len(items))
	for _, item := range items {
		if opsMatch(item, q) {
			filtered = append(filtered, item)
		}
	}
	sortOpsRecords(filtered)
	q = opsQueryDefaults(q)
	start, end := opsBounds(len(filtered), q.Page, q.PageSize)
	return OpsPage{Items: filtered[start:end], Page: q.Page, PageSize: q.PageSize, Total: len(filtered), HasNext: end < len(filtered)}, nil
}
func (s *OpsService) Transition(ctx context.Context, id string, expected int, target OpsStatus, actor string) (OpsRecord, error) {
	ctx, cancel := opsContext(ctx, 3*time.Second)
	defer cancel()
	// Honor request cancellation before performing any work. If the operator
	// already hung up, we must not touch the record or the audit log at all.
	if err := ctx.Err(); err != nil {
		return OpsRecord{}, err
	}
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return OpsRecord{}, err
	}
	if expected > 0 && expected != record.Revision {
		return OpsRecord{}, ErrOpsConflict
	}
	from := record.Status
	if !s.state.CanMove(from, target) {
		return OpsRecord{}, fmt.Errorf("%w: %s to %s", ErrOpsTransition, from, target)
	}
	record.Status = target
	if err := opsDelay(ctx, 250*time.Millisecond); err != nil {
		// The request was cancelled (client disconnect or timeout) while we
		// were waiting. Bail out before persisting the status change.
		return OpsRecord{}, err
	}
	// Re-check after the delay: a cancel that arrived during the wait must
	// stop the transition from landing in the store or the audit log.
	if err := ctx.Err(); err != nil {
		return OpsRecord{}, err
	}
	// Record the transition in the state machine only once we are committed to
	// persisting it, so a cancelled request leaves no trace behind.
	if err := s.state.Move(from, target, "operator update"); err != nil {
		return OpsRecord{}, err
	}
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.AddCtx(ctx, record.ID, "status_changed", actor)
	return record, nil
}
func (s *OpsService) Audit(id string) []OpsEvent { return s.audit.For(id) }
func (s *OpsService) Snapshot() OpsSnapshot {
	items, _ := s.store.List(context.Background())
	out := OpsSnapshot{Domain: opsDomainName, GeneratedAt: s.clock.Stamp(), ByStatus: map[OpsStatus]int{}, ByPriority: map[OpsPriority]int{}}
	for _, i := range items {
		out.Records++
		out.ByStatus[i.Status]++
		out.ByPriority[i.Priority]++
		if i.Status == OpsStatusActive {
			out.Active++
		}
	}
	return out
}
func (s *OpsService) Domain() string { return opsDomainName }
func (s *OpsService) Count() int     { return s.store.Count() }
func timeNowOps() string             { return time.Now().UTC().Format(time.RFC3339Nano) }
