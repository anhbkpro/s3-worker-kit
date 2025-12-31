package otel

import (
	"context"

	"s3-worker-kit/internal/config"
	"s3-worker-kit/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracerProvider initializes the OpenTelemetry tracer provider with the given configuration
func InitTracerProvider(cfg config.Config, logger observability.Logger) *sdktrace.TracerProvider {
	otelCfg := cfg.Otel

	if !otelCfg.Enabled {
		logger.Info("tracing disabled, using no-op tracer provider")
		// Create resource with service name even for disabled tracing
		res, err := resource.New(
			context.Background(),
			resource.WithAttributes(
				semconv.ServiceNameKey.String(observability.ServiceName),
			),
		)
		if err != nil {
			logger.Error("failed to create resource", "error", err)
			res = resource.Default()
		}

		// Return a no-op tracer provider if OTEL is disabled
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		return tp
	}

	logger.Info("initializing tracer provider",
		"exporter", otelCfg.Exporter,
		"endpoint", otelCfg.Endpoint,
		"insecure", otelCfg.Insecure,
	)

	// Create resource with service name
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(observability.ServiceName),
		),
	)
	if err != nil {
		logger.Error("failed to create resource", "error", err)
		res = resource.Default()
	}

	var spanExporter sdktrace.SpanExporter

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
		sdktrace.WithResource(res),
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
	logger.Info("tracer provider initialized and set as global provider",
		"service_name", observability.ServiceName)

	return tp
}
