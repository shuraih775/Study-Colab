package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

var ChatMessagesSent metric.Int64Counter

func InitChatMetrics(activeConnFn func() int) error {
	var err error

	ChatMessagesSent, err = Meter.Int64Counter(
		"chat.messages.sent_total",
	)
	if err != nil {
		return err
	}

	_, err = Meter.Int64ObservableGauge(
		"chat.connections.active",
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(activeConnFn()))
			return nil
		}),
	)
	return err
}
