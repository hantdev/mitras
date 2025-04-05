package consumers

import (
	"errors"

	"github.com/hantdev/mitras/pkg/messaging"
)

// ErrNotify wraps sending notification errors.
var ErrNotify = errors.New("error sending notification")

// Notifier represents an API for sending notification.
//
//go:generate mockery --name Notifier --output=./mocks --filename notifier.go --quiet
type Notifier interface {
	// Notify method is used to send notification for the
	// received message to the provided list of receivers.
	Notify(from string, to []string, msg *messaging.Message) error
}
