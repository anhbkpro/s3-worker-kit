package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracerProvider defines the interface for getting tracers
type TracerProvider = oteltrace.TracerProvider

// TraceMiddleware creates a Gin middleware for distributed tracing
func TraceMiddleware(serviceName string, tp TracerProvider) gin.HandlerFunc {
	if tp == nil {
		// Check if global tracer provider is set
		tp = otel.GetTracerProvider()
		if tp == nil {
			// Initialize a simple tracer provider for development/testing
			tp = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
			)
			// Set it as global so other parts of the app can use it
			otel.SetTracerProvider(tp)
		}
	}

	tracer := tp.Tracer("s3-worker-kit", oteltrace.WithInstrumentationVersion("v1.0.0"))
	return func(c *gin.Context) {
		// Extract the span context from incoming request headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span
		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPURL(c.Request.URL.String()),
				semconv.HTTPUserAgent(c.Request.UserAgent()),
				semconv.HTTPScheme(c.Request.URL.Scheme),
			),
		)
		defer span.End()

		// Add trace ID to response headers for debugging
		traceID := "no-span"
		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}
		c.Header("X-Trace-ID", traceID)
		// Also store in context for use in handlers
		c.Set("trace_id", traceID)
		c.Set("span", span)

		// Store span in context for use in handlers
		c.Request = c.Request.WithContext(ctx)

		// Wrap the response writer to capture status code
		c.Next()

		// Update span with response information
		span.SetAttributes(
			semconv.HTTPStatusCode(c.Writer.Status()),
		)

		// Mark span as error if status code indicates an error
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
		} else {
			span.SetStatus(codes.Ok, "HTTP request completed")
		}
	}
}

// GetTraceIDFromContext extracts trace ID from Gin context
func GetTraceIDFromContext(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if tid, ok := traceID.(string); ok {
			return tid
		}
	}
	return ""
}

// GetSpanFromContext extracts span from Gin context
func GetSpanFromContext(c *gin.Context) oteltrace.Span {
	if span, exists := c.Get("span"); exists {
		if s, ok := span.(oteltrace.Span); ok {
			return s
		}
	}
	return nil
}

// CreateSpan creates a child span for internal operations
func CreateSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return otel.Tracer("s3-worker-kit").Start(ctx, name, opts...)
}

// AddSpanAttributes adds attributes to the current span from context
func AddSpanAttributes(ctx context.Context, attrs ...oteltrace.SpanStartOption) {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		// Note: This is a simplified version. In practice, you'd want to convert
		// the span start options to span attributes
		for _, opt := range attrs {
			// This is a placeholder - you'd need to implement proper attribute setting
			_ = opt
		}
	}
}
