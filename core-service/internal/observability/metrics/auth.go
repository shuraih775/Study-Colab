package metrics

import "go.opentelemetry.io/otel/metric"

var (
	AuthLoginSuccess metric.Int64Counter
	AuthLoginFailure metric.Int64Counter
)

func InitAuthMetrics() error {
	var err error

	AuthLoginSuccess, err = Meter.Int64Counter(
		"auth.login.success_total",
	)
	if err != nil {
		return err
	}

	AuthLoginFailure, err = Meter.Int64Counter(
		"auth.login.failure_total",
	)
	return err
}
