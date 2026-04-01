package logger_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/weprodev/go-pkg/logger"
)

func TestLogger_NewWithWriter(t *testing.T) {
	var buf bytes.Buffer
	cfg := logger.Config{
		Level:  logger.LevelInfo,
		Format: logger.FormatJSON,
	}

	log, err := logger.NewWithWriter(&buf, cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	log.Info("hello world", "key", "value")

	output := buf.String()
	if !contains(output, "hello world") || !contains(output, "value") {
		t.Errorf("log output didn't contain expected strings: %s", output)
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	extractor := func(ctx context.Context) []any {
		if val, ok := ctx.Value("trace_id").(string); ok {
			return []any{"trace_id", val}
		}
		return nil
	}

	cfg := logger.Config{
		Level:             logger.LevelDebug,
		Format:            logger.FormatText,
		ContextExtractors: []logger.ContextExtractor{extractor},
	}

	log, _ := logger.NewWithWriter(&buf, cfg)
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")

	logCtx := log.WithContext(ctx)
	logCtx.Debug("context matters")

	output := buf.String()
	if !contains(output, "abc-123") {
		t.Errorf("context trace_id was not logged: %s", output)
	}
}

func TestLogger_WithErrorTime(t *testing.T) {
	var buf bytes.Buffer
	cfg := logger.Config{Format: logger.FormatText}
	log, _ := logger.NewWithWriter(&buf, cfg)

	// Should be a no-op (and must not panic).
	log = log.WithError(nil)

	testErr := errors.New("something broke")
	testTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	log.WithError(testErr).WithTime(testTime).Error("failure")

	output := buf.String()
	if !contains(output, "something broke") {
		t.Errorf("error was not logged: %s", output)
	}
	if !contains(output, "2025-01") {
		t.Errorf("time was not logged: %s", output)
	}
}

func TestLogger_Close_NoOpForWriterLogger(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.NewWithWriter(&buf, logger.Config{})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestLogger_MultiHandler(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	h1 := slog.NewTextHandler(&buf1, nil)
	h2 := slog.NewJSONHandler(&buf2, nil)

	cfg := logger.Config{
		ExtraHandlers: []slog.Handler{h1, h2},
	}
	log, _ := logger.NewWithWriter(&bytes.Buffer{}, cfg)

	log.Info("fanning out")

	if !contains(buf1.String(), "fanning out") {
		t.Errorf("h1 missed log")
	}
	if !contains(buf2.String(), "fanning out") {
		t.Errorf("h2 missed log")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
