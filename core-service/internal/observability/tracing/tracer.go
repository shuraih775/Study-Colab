package tracing

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Config struct {
	ServiceName string
	Endpoint    string
	Exporter    sdktrace.SpanExporter
	Sampler     sdktrace.Sampler
}

var initOnce = &sync.Once{}

func Init(cfg Config) (func(context.Context) error, error) {
	var tp *sdktrace.TracerProvider
	var initErr error

	initOnce.Do(func() {
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)

		var exporter sdktrace.SpanExporter
		if cfg.Exporter != nil {
			exporter = cfg.Exporter
		} else {
			exp, err := otlptracegrpc.New(
				context.Background(),
				otlptracegrpc.WithEndpoint(cfg.Endpoint),
				otlptracegrpc.WithInsecure(),
			)
			if err != nil {
				initErr = err
				return
			}
			exporter = exp
		}

		sampler := cfg.Sampler
		if sampler == nil {
			sampler = sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(1.0),
			)
		}

		res, err := resource.New(
			context.Background(),
			resource.WithTelemetrySDK(),
			resource.WithHost(),
			resource.WithAttributes(
				semconv.ServiceName(cfg.ServiceName),
			),
		)
		if err != nil {
			initErr = err
			return
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(
				exporter,
				sdktrace.WithBatchTimeout(5*time.Second),
			),
			sdktrace.WithSampler(sampler),
			sdktrace.WithResource(res),
		)

		otel.SetTracerProvider(tp)
	})

	if initErr != nil {
		return nil, initErr
	}

	return tp.Shutdown, nil
}
