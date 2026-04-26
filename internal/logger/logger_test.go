package logger_test

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/tarotclub/go-tpl/internal/logger"
)

func TestNew(t *testing.T) {
	tests := []struct {
		level    string
		wantLvl  zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"invalid", zerolog.InfoLevel}, // falls back to info
	}

	for _, tc := range tests {
		log := logger.New(tc.level)
		if log.GetLevel() != tc.wantLvl {
			t.Errorf("New(%q).GetLevel() = %v; want %v", tc.level, log.GetLevel(), tc.wantLvl)
		}
	}
}
