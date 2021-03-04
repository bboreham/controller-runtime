package tracing

import (
	"context"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlphttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// SetupOTLP sets up a global trace provider sending to OpenTelemetry with some defaults
func SetupOTLP(serviceName string) (io.Closer, error) {
	ctx := context.Background()
	driver := otlphttp.NewDriver(
		otlphttp.WithInsecure(),
		otlphttp.WithEndpoint("otlp-collector.default:55680"),
	)
	exp, err := otlp.NewExporter(context.Background(), driver)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	return otlpCloser{exp: exp, bsp: bsp}, nil
}

type otlpCloser struct {
	exp *otlp.Exporter
	bsp *sdktrace.BatchSpanProcessor
}

func (s otlpCloser) Close() error {
	s.bsp.Shutdown(context.Background()) // shutdown the processor
	return s.exp.Shutdown(context.Background())
}
