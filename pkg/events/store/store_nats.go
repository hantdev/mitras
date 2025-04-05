//go:build nats
// +build nats

package store

import (
	"context"
	"log"
	"log/slog"

	"github.com/hantdev/mitras/pkg/events"
	"github.com/hantdev/mitras/pkg/events/nats"
)

// StreamAllEvents represents subject to subscribe for all the events.
const StreamAllEvents = "events.>"

func init() {
	log.Println("The binary was build using nats as the events store")
}

func NewPublisher(ctx context.Context, url string) (events.Publisher, error) {
	pb, err := nats.NewPublisher(ctx, url)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewSubscriber(ctx context.Context, url string, logger *slog.Logger) (events.Subscriber, error) {
	pb, err := nats.NewSubscriber(ctx, url, logger)
	if err != nil {
		return nil, err
	}

	return pb, nil
}
