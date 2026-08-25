package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

// 并发 summary 请求与缓存读写不得产生 data race。
func TestPassSummaryConcurrentCacheRace(t *testing.T) {
	store := NewPassStore()
	ctx := context.Background()
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		_, err := store.Add(ctx, PassWindow{Satellite: "Beacon", Station: "West Mesa", Start: now, End: now.Add(time.Hour), State: "planned"})
		if err != nil {
			t.Fatal(err)
		}
	}
	stats := newPassStats()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 30; j++ {
				_, _ = stats.Summary(ctx, store)
				_ = stats.Cached()
			}
		}()
	}
	close(start)
	wg.Wait()
}

// ticker 收到 stop 信号后必须退出。
func TestPassTickerStopsOnSignal(t *testing.T) {
	store := NewPassStore()
	stats := newPassStats()
	stop := make(chan struct{})
	done := startStatsTicker(stats, store, 10*time.Millisecond, stop)
	close(stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stats ticker goroutine did not stop after signal")
	}
}

// ticker 必须周期性反复刷新缓存。
func TestPassTickerRefreshesCacheRepeatedly(t *testing.T) {
	store := NewPassStore()
	stats := newPassStats()
	ctx := context.Background()
	stop := make(chan struct{})
	done := startStatsTicker(stats, store, 15*time.Millisecond, stop)
	defer func() {
		close(stop)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("ticker goroutine leaked")
		}
	}()

	now := time.Now().UTC()
	_, _ = store.Add(ctx, PassWindow{Satellite: "Beacon", Station: "West Mesa", Start: now, End: now.Add(time.Hour), State: "planned"})
	time.Sleep(80 * time.Millisecond)
	_, _ = store.Add(ctx, PassWindow{Satellite: "Beacon-2", Station: "North Ridge", Start: now, End: now.Add(time.Hour), State: "planned"})
	time.Sleep(80 * time.Millisecond)

	if cached := stats.Cached(); cached.Total != 3 {
		t.Fatalf("ticker did not refresh repeatedly: cached total = %d, want 3", cached.Total)
	}
}

// interval 非法（<=0）时 ticker 必须用默认间隔启动，不能 panic。
func TestPassTickerHandlesZeroInterval(t *testing.T) {
	store := NewPassStore()
	stats := newPassStats()
	stop := make(chan struct{})
	done := startStatsTicker(stats, store, 0, stop)
	close(stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ticker with zero interval did not stop")
	}
}
