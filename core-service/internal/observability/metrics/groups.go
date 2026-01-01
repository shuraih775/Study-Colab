package metrics

import "go.opentelemetry.io/otel/metric"

var (
	GroupsCreated     metric.Int64Counter
	GroupsJoined      metric.Int64Counter
	GroupsJoinRequest metric.Int64Counter
)

func InitGroupMetrics() error {
	var err error

	GroupsCreated, err = Meter.Int64Counter(
		"groups.created_total",
	)
	if err != nil {
		return err
	}

	GroupsJoined, err = Meter.Int64Counter(
		"groups.joined_total",
	)
	if err != nil {
		return err
	}

	GroupsJoinRequest, err = Meter.Int64Counter(
		"groups.join_requests_total",
	)
	return err
}
