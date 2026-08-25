package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrExportPoolExhausted = errors.New("pass export lease pool exhausted")

const passExportLeaseLimit = 8

type PassExporter struct {
	store  *PassStore
	audit  *PassAudit
	tokens chan struct{}
}

func newPassExporter(store *PassStore, audit *PassAudit) *PassExporter {
	return &PassExporter{store: store, audit: audit, tokens: make(chan struct{}, passExportLeaseLimit)}
}

func (e *PassExporter) acquire() (release func(), ok bool) {
	select {
	case e.tokens <- struct{}{}:
		var once bool
		return func() {
			if !once {
				once = true
				<-e.tokens
			}
		}, true
	default:
		return nil, false
	}
}

func (e *PassExporter) finalize(result ExportPassResult) error {
	if result.Exported == 0 {
		return nil
	}
	for _, id := range result.IDs {
		e.audit.Add(id, "exported", "system")
	}
	return nil
}

// Export moves every window currently in state `from` into state `to`.
func (e *PassExporter) Export(ctx context.Context, from, to string) (result ExportPassResult, err error) {
	if !passStateValid(from) || !passStateValid(to) || from == to {
		return result, fmt.Errorf("%w: from %s to %s", ErrPassInvalid, from, to)
	}
	items, err := e.store.List(ctx)
	if err != nil {
		return result, err
	}
	result = ExportPassResult{IDs: []string{}}
	ids := make(chan string, len(items))
	errCh := make(chan error, 1)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, item := range items {
		if item.State != from {
			result.Skipped++
			continue
		}
		wg.Add(1)
		go func(item PassWindow) {
			defer wg.Done()
			<-start
			release, ok := e.acquire()
			if !ok {
				select {
				case errCh <- fmt.Errorf("%w after %d windows", ErrExportPoolExhausted, result.Exported):
				default:
				}
				return
			}
			defer release()
			updated, err := e.store.UpdateState(ctx, item.ID, to)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			ids <- updated.ID
		}(item)
	}
	close(start)
	wg.Wait()
	close(ids)
	for id := range ids {
		result.IDs = append(result.IDs, id)
	}
	result.Exported = len(result.IDs)
	if err := e.finalize(result); err != nil {
		return result, err
	}
	return result, nil
}
