package channels

import "errors"

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid channels status")

	// ErrEnableChannel indicates error in enabling channel.
	ErrEnableChannel = errors.New("failed to enable channel")

	// ErrDisableChannel indicates error in disabling channel.
	ErrDisableChannel = errors.New("failed to disable channel")
)
