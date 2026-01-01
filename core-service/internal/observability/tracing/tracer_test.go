package tracing

import (
	"context"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func resetForTest() {
	initOnce = &sync.Once{}
}

func TestInit_UsesInjectedExporter(t *testing.T) {
	resetForTest()

	exporter := tracetest.NewInMemoryExporter()

	shutdown, err := Init(Config{
		ServiceName: "test-service",
		Exporter:    exporter,
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer shutdown(context.Background())

	tracer := otel.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	tp := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	if err := tp.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush failed: %v", err)
	}

	if len(exporter.GetSpans()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(exporter.GetSpans()))
	}
}

func TestInit_SamplerDropsSpans(t *testing.T) {
	resetForTest()

	exporter := tracetest.NewInMemoryExporter()

	shutdown, err := Init(Config{
		ServiceName: "test-service",
		Exporter:    exporter,
		Sampler:     sdktrace.NeverSample(),
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer shutdown(context.Background())

	tracer := otel.Tracer("test")

	_, span := tracer.Start(context.Background(), "should-not-exist")
	span.End()

	tp := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	_ = tp.ForceFlush(context.Background())

	if len(exporter.GetSpans()) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(exporter.GetSpans()))
	}
}
