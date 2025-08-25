package loggerpkg

import (
	"bytes"
	"log/slog"
)

type Logger struct {
	buf    bytes.Buffer
	logger *slog.Logger
}

func NewLogger() *Logger {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &Logger{buf: buf, logger: logger}
}
