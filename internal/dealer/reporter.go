package dealer

import "time"

// there are various metrics implementation variations, depending on how varied are the expected implementation.
// One common implementation is to have a single goroutine to record the metrics. This goroutine Will then call a Reporter interface that
// can be provided when the agent starts. In Go, the preferred convention for implementing the Reporter is to use a middleware.
// Another solution would be the usage of static functions.

type Metric int

const (
	// SubmitOrderMetric Submit order metrics.
	SubmitOrderMetric Metric = iota
	SubmitOrderLatencyMetric
	SubmitOrderErrorMetric
	// SubmitBulkOrderMetric Submit bulk orders metrics.
	SubmitBulkOrderMetric
	SubmitBulkOrderLatencyMetric
	// ModifyOrderMetric Modify order metrics.
	ModifyOrderMetric
	ModifyOrderLatencyMetric
	ModifyOrderErrorMetric
	// CancelOrderMetric Cancel order metrics.
	CancelOrderMetric
	CancelOrderLatencyMetric
	CancelOrderErrorMetric
	// CancelAllOrdersMetric Cancel all orders metrics.
	CancelAllOrdersMetric
	CancelAllOrdersLatencyMetric
	CancelAllOrdersErrorMetric
	// GetActiveOrdersMetric Get Active orders metrics.
	GetActiveOrdersMetric
	GetActiveOrdersLatencyMetric
	GetActiveOrdersErrorMetric
	// MaxMetrics this should always be the last one.
	MaxMetrics
)

// Reporter interface is implemented by the clients that wish to report their metrics for this library.
type Reporter interface {
	// Event metrics in a single occurrence
	Event(m Metric, labels ...string)
	// Latency metrics provide visibility over latencies
	Latency(m Metric, d time.Duration, labels ...string)
	// Value metrics provide visibility over arbitrary value
	Value(m Metric, v float64, labels ...string)
}
