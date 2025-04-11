package middleware

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/hantdev/mitras/journal"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
)

var _ journal.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	service journal.Service
}

// MetricsMiddleware returns new message repository
// with Save method wrapped to expose metrics.
func MetricsMiddleware(service journal.Service, counter metrics.Counter, latency metrics.Histogram) journal.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		service: service,
	}
}

func (mm *metricsMiddleware) Save(ctx context.Context, j journal.Journal) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "save").Add(1)
		mm.latency.With("method", "save").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.service.Save(ctx, j)
}

func (mm *metricsMiddleware) RetrieveAll(ctx context.Context, session smqauthn.Session, page journal.Page) (journal.JournalsPage, error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "retrieve_all").Add(1)
		mm.latency.With("method", "retrieve_all").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.service.RetrieveAll(ctx, session, page)
}
