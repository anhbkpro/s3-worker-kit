package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

const ServiceName = "s3-worker-kit"

// Logger defines the logging interface
type Logger interface {
	With(args ...any) Logger
	WithGroup(name string) Logger
	WithContext(ctx context.Context) Logger

	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

// SlogLogger implements Logger using slog
type SlogLogger struct {
	logger *slog.Logger
	attrs  []slog.Attr
}

// NewLogger creates a new structured logger with JSON output and service name
func NewLogger(serviceName string, output io.Writer) Logger {
	if output == nil {
		output = os.Stdout
	}

	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add timestamp in RFC3339 format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	})

	logger := slog.New(handler)

	return &SlogLogger{
		logger: logger,
		attrs: []slog.Attr{
			slog.String("service_name", serviceName),
		},
	}
}

// NewDefaultLogger creates a logger with the default service name
func NewDefaultLogger() Logger {
	return NewLogger(ServiceName, os.Stdout)
}

func (l *SlogLogger) With(args ...any) Logger {
	attrs := make([]slog.Attr, len(l.attrs))
	copy(attrs, l.attrs)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := args[i].(string)
			value := args[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}

	return &SlogLogger{
		logger: l.logger,
		attrs:  attrs,
	}
}

func (l *SlogLogger) WithGroup(name string) Logger {
	return &SlogLogger{
		logger: l.logger.WithGroup(name),
		attrs:  l.attrs,
	}
}

func (l *SlogLogger) WithContext(ctx context.Context) Logger {
	return &SlogLogger{
		logger: l.logger,
		attrs:  l.attrs,
	}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *SlogLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(level, msg, args...)
}

func (l *SlogLogger) log(level slog.Level, msg string, args ...any) {
	var slogArgs []any

	// Convert attrs to args
	for _, attr := range l.attrs {
		slogArgs = append(slogArgs, attr.Key, attr.Value.Any())
	}

	// Add additional args
	slogArgs = append(slogArgs, args...)

	l.logger.Log(context.Background(), level, msg, slogArgs...)
}
