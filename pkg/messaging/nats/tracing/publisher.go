package tracing

import (
	"context"

	"github.com/hantdev/mitras/pkg/messaging"
	"github.com/hantdev/mitras/pkg/messaging/tracing"
	"github.com/hantdev/mitras/pkg/server"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Traced operations.
const publishOP = "publish"

var defaultAttributes = []attribute.KeyValue{
	attribute.String("messaging.system", "nats"),
	attribute.String("network.protocol.name", "nats"),
	attribute.String("network.protocol.version", "2.2.4"),
}

var _ messaging.Publisher = (*publisherMiddleware)(nil)

type publisherMiddleware struct {
	publisher messaging.Publisher
	tracer    trace.Tracer
	host      server.Config
}

func NewPublisher(config server.Config, tracer trace.Tracer, publisher messaging.Publisher) messaging.Publisher {
	pub := &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
		host:      config,
	}

	return pub
}

func (pm *publisherMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	ctx, span := tracing.CreateSpan(ctx, publishOP, msg.GetPublisher(), topic, msg.GetSubtopic(), len(msg.GetPayload()), pm.host, trace.SpanKindClient, pm.tracer)
	defer span.End()
	span.SetAttributes(defaultAttributes...)

	return pm.publisher.Publish(ctx, topic, msg)
}

func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}
