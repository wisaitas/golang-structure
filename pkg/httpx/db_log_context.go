package httpx

import (
	"context"
	"sync"
)

type dbLogContextKey struct{}

type dbLogCollector struct {
	mu   sync.Mutex
	logs []DBLog
}

func WithDBLogCollector(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, dbLogContextKey{}, &dbLogCollector{logs: make([]DBLog, 0, 8)})
}

func AddDBLog(ctx context.Context, dbLog DBLog) {
	if ctx == nil {
		return
	}
	collector, ok := ctx.Value(dbLogContextKey{}).(*dbLogCollector)
	if !ok || collector == nil {
		return
	}

	collector.mu.Lock()
	collector.logs = append(collector.logs, dbLog)
	collector.mu.Unlock()
}

func GetDBLogs(ctx context.Context) []DBLog {
	if ctx == nil {
		return nil
	}
	collector, ok := ctx.Value(dbLogContextKey{}).(*dbLogCollector)
	if !ok || collector == nil {
		return nil
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	if len(collector.logs) == 0 {
		return nil
	}

	result := make([]DBLog, len(collector.logs))
	copy(result, collector.logs)
	return result
}
