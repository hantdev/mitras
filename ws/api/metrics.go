//go:build !test

package api

import (
	"context"
	"time"

	"github.com/hantdev/mitras/ws"
	"github.com/go-kit/kit/metrics"
)

var _ ws.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     ws.Service
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc ws.Service, counter metrics.Counter, latency metrics.Histogram) ws.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

// Subscribe instruments Subscribe method with metrics.
func (mm *metricsMiddleware) Subscribe(ctx context.Context, clientKey, chanID, subtopic string, c *ws.Client) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "subscribe").Add(1)
		mm.latency.With("method", "subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Subscribe(ctx, clientKey, chanID, subtopic, c)
}