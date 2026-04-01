package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// Level represents the logging level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format represents the logging format
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// ContextExtractor defines a function that extracts logging attributes from a context
type ContextExtractor func(ctx context.Context) []any

// Config holds the logger configuration
type Config struct {
	Level             Level
	Format            Format
	OutputPath        string
	ExtraHandlers     []slog.Handler     // Provider agnostic hook injection (Sentry, OpenTelemetry, etc.)
	ContextExtractors []ContextExtractor // Functions to extract custom tags/logs from Context
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	extractors []ContextExtractor
	closer     io.Closer
}

// New creates a new logger instance
func New(config Config) (*Logger, error) {
	output := io.Writer(os.Stdout)
	var closer io.Closer
	if config.OutputPath != "" && config.OutputPath != "stdout" {
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		output = file
		closer = file
	}
	log, err := NewWithWriter(output, config)
	if err != nil {
		if closer != nil {
			_ = closer.Close()
		}
		return nil, err
	}
	log.closer = closer
	return log, nil
}

// NewWithWriter creates a new logger instance with a custom writer
func NewWithWriter(output io.Writer, config Config) (*Logger, error) {
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: level,
		})
	case FormatText:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: level,
		})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: level,
		})
	}

	// Provider agnostic MultiHandler pattern
	handlers := []slog.Handler{handler}
	if len(config.ExtraHandlers) > 0 {
		handlers = append(handlers, config.ExtraHandlers...)
	}

	var finalHandler slog.Handler
	if len(handlers) == 1 {
		finalHandler = handlers[0]
	} else {
		finalHandler = &multiHandler{handlers: handlers}
	}

	return &Logger{
		Logger:     slog.New(finalHandler),
		extractors: config.ContextExtractors,
	}, nil
}

// Close closes any underlying file handle opened by New.
// It is a no-op for loggers created via NewWithWriter.
func (l *Logger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}

// multiHandler fans out slog records to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs[0] // Return first error if any fail
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// WithContext adds context values to the logger using configured extractors
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if len(l.extractors) == 0 {
		return l
	}

	var args []any
	for _, extractor := range l.extractors {
		if extracted := extractor(ctx); len(extracted) > 0 {
			args = append(args, extracted...)
		}
	}

	if len(args) == 0 {
		return l
	}

	return &Logger{
		Logger:     l.With(args...),
		extractors: l.extractors,
		closer:     l.closer,
	}
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		Logger:     l.With("error", err.Error()),
		extractors: l.extractors,
		closer:     l.closer,
	}
}

// WithTime adds a timestamp to the logger
func (l *Logger) WithTime(t time.Time) *Logger {
	return &Logger{
		Logger:     l.With("time", t.Format(time.RFC3339)),
		extractors: l.extractors,
		closer:     l.closer,
	}
}
