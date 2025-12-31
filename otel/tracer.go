package otel

import (
	"context"

	"s3-worker-kit/internal/config"
	"s3-worker-kit/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracerProvider initializes the OpenTelemetry tracer provider with the given configuration
func InitTracerProvider(cfg config.Config, logger observability.Logger) *sdktrace.TracerProvider {
	otelCfg := cfg.Otel

	if !otelCfg.Enabled {
		logger.Info("tracing disabled, using no-op tracer provider")
		// Return a no-op tracer provider if OTEL is disabled
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		return tp
	}

	logger.Info("initializing tracer provider",
		"exporter", otelCfg.Exporter,
		"endpoint", otelCfg.Endpoint,
		"insecure", otelCfg.Insecure,
	)

	var spanExporter sdktrace.SpanExporter
	var err error

	switch otelCfg.Exporter {
	case "otlp":
		// Create OTLP gRPC exporter
		logger.Info("creating OTLP gRPC exporter", "endpoint", otelCfg.Endpoint)
		spanExporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(otelCfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			logger.Error("failed to create OTLP exporter, falling back to no exporter",
				"error", err,
				"endpoint", otelCfg.Endpoint,
			)
			spanExporter = nil
		} else {
			logger.Info("OTLP gRPC exporter created successfully", "endpoint", otelCfg.Endpoint)
		}
	default:
		logger.Warn("unsupported exporter, using no exporter", "exporter", otelCfg.Exporter)
		spanExporter = nil
	}

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}

	if spanExporter != nil {
		tpOpts = append(tpOpts, sdktrace.WithBatcher(spanExporter))
		logger.Info("span batcher configured for exporter")
	} else {
		logger.Info("no span exporter configured, traces will be stored in memory only")
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)
	logger.Info("tracer provider initialized and set as global provider")

	return tp
}
