package webserver

import (
	"exp_poc/metrics"
	"time"
)

type MetricsHandler struct {
	TimeoutMargin    float64
	CollectorFactory func(timeout time.Duration, requestedCollectors []string) (error, *metrics.CustomCollector)
}
