package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracerProvider initializes the OpenTelemetry tracer provider with the given configuration
func InitTracerProvider(enabled bool, exporter, endpoint string, insecure bool) *sdktrace.TracerProvider {
	if !enabled {
		log.Println("otel: tracing disabled, using no-op tracer provider")
		// Return a no-op tracer provider if OTEL is disabled
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		return tp
	}

	log.Printf("otel: initializing tracer provider with exporter=%s, endpoint=%s, insecure=%t", exporter, endpoint, insecure)

	var spanExporter sdktrace.SpanExporter
	var err error

	switch exporter {
	case "otlp":
		// Create OTLP gRPC exporter
		log.Printf("otel: creating OTLP gRPC exporter to %s", endpoint)
		spanExporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			log.Printf("otel: failed to create OTLP exporter: %v, falling back to no exporter", err)
			spanExporter = nil
		} else {
			log.Println("otel: OTLP gRPC exporter created successfully")
		}
	default:
		log.Printf("otel: unsupported exporter '%s', using no exporter", exporter)
		spanExporter = nil
	}

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}

	if spanExporter != nil {
		tpOpts = append(tpOpts, sdktrace.WithBatcher(spanExporter))
		log.Println("otel: span batcher configured for exporter")
	} else {
		log.Println("otel: no span exporter configured, traces will be stored in memory only")
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)
	log.Println("otel: tracer provider initialized and set as global provider")

	return tp
}
