package athena

import "context"

// MessagePublisher specifies a message publishing API.
type MessagePublisher interface {
	// Publishes message to the stream. A non-nil error is returned to indicate
	// operation failure.
	Publish(context.Context, string, Message) error
}
