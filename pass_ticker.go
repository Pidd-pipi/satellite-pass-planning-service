package main

import (
	"context"
	"time"
)

// startStatsTicker 周期性重算统计缓存，返回的 done 通道在 ticker 停止后关闭。
func startStatsTicker(stats *PassStats, store *PassStore, interval time.Duration, stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		t := time.NewTimer(interval)
		for {
			select {
			case <-t.C:
				_, _ = stats.Summary(context.Background(), store)
			}
		}
		_ = done
	}()
	return done
}
