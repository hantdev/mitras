package events

import (
	"context"

	"github.com/hantdev/mitras/pkg/events"
	"github.com/hantdev/mitras/pkg/events/store"
)

const streamID = "mitras.mqtt"

//go:generate mockery --name EventStore --output=../mocks --filename events.go --quiet
type EventStore interface {
	Connect(ctx context.Context, clientID string) error
	Disconnect(ctx context.Context, clientID string) error
}

// EventStore is a struct used to store event streams in Redis.
type eventStore struct {
	events.Publisher
	instance string
}

// NewEventStore returns wrapper around mProxy service that sends
// events to event store.
func NewEventStore(ctx context.Context, url, instance string) (EventStore, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		instance:  instance,
		Publisher: publisher,
	}, nil
}

// Connect issues event on MQTT CONNECT.
func (es *eventStore) Connect(ctx context.Context, clientID string) error {
	ev := mqttEvent{
		clientID:  clientID,
		operation: "connect",
		instance:  es.instance,
	}

	return es.Publish(ctx, ev)
}

// Disconnect issues event on MQTT CONNECT.
func (es *eventStore) Disconnect(ctx context.Context, clientID string) error {
	ev := mqttEvent{
		clientID:  clientID,
		operation: "disconnect",
		instance:  es.instance,
	}

	return es.Publish(ctx, ev)
}
