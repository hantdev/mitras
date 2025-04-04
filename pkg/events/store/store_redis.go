//go:build !nats && !rabbitmq
// +build !nats,!rabbitmq

package store

import (
	"context"
	"log"
	"log/slog"

	"github.com/hantdev/mitras/pkg/events"
	"github.com/hantdev/mitras/pkg/events/redis"
)

// StreamAllEvents represents subject to subscribe for all the events.
const StreamAllEvents = ">"

func init() {
	log.Println("The binary was build using redis as the events store")
}

func NewPublisher(ctx context.Context, url string) (events.Publisher, error) {
	pb, err := redis.NewPublisher(ctx, url, events.UnpublishedEventsCheckInterval)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewSubscriber(_ context.Context, url string, logger *slog.Logger) (events.Subscriber, error) {
	pb, err := redis.NewSubscriber(url, logger)
	if err != nil {
		return nil, err
	}

	return pb, nil
}
