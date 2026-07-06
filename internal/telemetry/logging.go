package telemetry

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type ctxKey string

const (
	traceIDKey ctxKey = "trace_id"
	spanIDKey  ctxKey = "span_id"
)

func InitLogger(level, format, service string) zerolog.Logger {
	var w io.Writer = os.Stdout

	if format == "console" {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	return zerolog.New(w).
		Level(lvl).
		With().
		Timestamp().
		Str("service", service).
		Logger()
}

func WithTraceContext(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, spanIDKey, spanID)
	return ctx
}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

func SpanIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

func LoggerFromContext(ctx context.Context, base zerolog.Logger) zerolog.Logger {
	l := base.With()
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		l = l.Str("trace_id", traceID)
	}
	if spanID := SpanIDFromContext(ctx); spanID != "" {
		l = l.Str("span_id", spanID)
	}
	return l.Logger()
}
