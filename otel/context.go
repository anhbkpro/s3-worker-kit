package otel

import "context"

type traceIDContextKey struct{}

var TraceIDKey = traceIDContextKey{}

func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func SetTraceIDInContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
