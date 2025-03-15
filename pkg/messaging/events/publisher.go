package events

import (
	"context"

	"github.com/hantdev/mitras/pkg/events"
	"github.com/hantdev/mitras/pkg/events/store"
	"github.com/hantdev/mitras/pkg/messaging"
)

var _ messaging.Publisher = (*publisherES)(nil)

type publisherES struct {
	ep  events.Publisher
	pub messaging.Publisher
}

func NewPublisherMiddleware(ctx context.Context, pub messaging.Publisher, url string) (messaging.Publisher, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}

	return &publisherES{
		ep:  publisher,
		pub: pub,
	}, nil
}

func (es *publisherES) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if err := es.pub.Publish(ctx, topic, msg); err != nil {
		return err
	}

	me := publishEvent{
		channelID: msg.Channel,
		clientID:  msg.Publisher,
		subtopic:  msg.Subtopic,
	}

	return es.ep.Publish(ctx, me)
}

func (es *publisherES) Close() error {
	return es.pub.Close()
}
