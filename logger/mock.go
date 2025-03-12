package logger

import (
	"bytes"
	"log/slog"
)

// NewMock returns wrapped slog logger mock.
func NewMock() *slog.Logger {
	buf := &bytes.Buffer{}

	return slog.New(slog.NewJSONHandler(buf, nil))
}
