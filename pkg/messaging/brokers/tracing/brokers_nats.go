//go:build !rabbitmq
// +build !rabbitmq

package brokers

import (
	"log"

	"github.com/hantdev/mitras/pkg/messaging"
	"github.com/hantdev/mitras/pkg/messaging/nats/tracing"
	"github.com/hantdev/mitras/pkg/server"
	"go.opentelemetry.io/otel/trace"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

func init() {
	log.Println("The binary was build using Nats as the message broker")
}

func NewPublisher(cfg server.Config, tracer trace.Tracer, publisher messaging.Publisher) messaging.Publisher {
	return tracing.NewPublisher(cfg, tracer, publisher)
}

func NewPubSub(cfg server.Config, tracer trace.Tracer, pubsub messaging.PubSub) messaging.PubSub {
	return tracing.NewPubSub(cfg, tracer, pubsub)
}