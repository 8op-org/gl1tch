package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Setup initialises the global OTel TracerProvider and MeterProvider.
// If OTEL_EXPORTER_OTLP_ENDPOINT is set, traces are exported via OTLP gRPC;
// otherwise they go to stdout. Metrics always go to stdout.
// The returned shutdown func must be called before the process exits.
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("dev"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build resource: %w", err)
	}

	traceExp, err := newTraceExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("telemetry: trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricExp, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("telemetry: metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExp)),
	)
	otel.SetMeterProvider(mp)

	shutdown := func(ctx context.Context) error {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
		return nil
	}
	return shutdown, nil
}

func newTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	}
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}
