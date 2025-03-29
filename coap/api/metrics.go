//go:build !test

package api

import (
	"context"
	"time"

	"github.com/hantdev/mitras/coap"
	"github.com/hantdev/mitras/pkg/messaging"
	"github.com/go-kit/kit/metrics"
)

var _ coap.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     coap.Service
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc coap.Service, counter metrics.Counter, latency metrics.Histogram) coap.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

// Publish instruments Publish method with metrics.
func (mm *metricsMiddleware) Publish(ctx context.Context, key string, msg *messaging.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "publish").Add(1)
		mm.latency.With("method", "publish").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Publish(ctx, key, msg)
}

// Subscribe instruments Subscribe method with metrics.
func (mm *metricsMiddleware) Subscribe(ctx context.Context, key, chanID, subtopic string, c coap.Client) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "subscribe").Add(1)
		mm.latency.With("method", "subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Subscribe(ctx, key, chanID, subtopic, c)
}

// Unsubscribe instruments Unsubscribe method with metrics.
func (mm *metricsMiddleware) Unsubscribe(ctx context.Context, key, chanID, subtopic, token string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "unsubscribe").Add(1)
		mm.latency.With("method", "unsubscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Unsubscribe(ctx, key, chanID, subtopic, token)
}

// DisconnectHandler instruments DisconnectHandler method with metrics.
func (mm *metricsMiddleware) DisconnectHandler(ctx context.Context, chanID, subtopic, token string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "disconnect_handler").Add(1)
		mm.latency.With("method", "disconnect_handler").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.DisconnectHandler(ctx, chanID, subtopic, token)
}