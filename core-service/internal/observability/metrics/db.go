package metrics

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel/metric"
)

func RegisterDBMetrics(db *sql.DB) error {
	_, err := Meter.Int64ObservableGauge(
		"db.connections.open",
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			stats := db.Stats()
			o.Observe(int64(stats.OpenConnections))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = Meter.Int64ObservableGauge(
		"db.connections.in_use",
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			stats := db.Stats()
			o.Observe(int64(stats.InUse))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = Meter.Int64ObservableGauge(
		"db.connections.idle",
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			stats := db.Stats()
			o.Observe(int64(stats.Idle))
			return nil
		}),
	)
	return err
}
