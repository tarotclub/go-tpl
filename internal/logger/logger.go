package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a zerolog logger with human-friendly console output and the given level.
func New(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).Level(lvl).With().Timestamp().Logger()
}
