package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reqCounter       = promauto.NewCounter(prometheus.CounterOpts{Name: "pvz_requests_total", Help: "Total HTTP requests"})
	latencyHist      = promauto.NewHistogram(prometheus.HistogramOpts{Name: "pvz_request_duration_seconds", Help: "Request latency in seconds", Buckets: prometheus.DefBuckets})
	pvzCounter       = promauto.NewCounter(prometheus.CounterOpts{Name: "pvz_created_total", Help: "Total PVZ created"})
	receptionCounter = promauto.NewCounter(prometheus.CounterOpts{Name: "receptions_created_total", Help: "Total receptions opened"})
	productCounter   = promauto.NewCounter(prometheus.CounterOpts{Name: "products_added_total", Help: "Total products added"})
)

type PromMetrics struct{}

func NewPromMetrics() *PromMetrics {
	return &PromMetrics{}
}

func (m *PromMetrics) IncRequest() {
	reqCounter.Inc()
}

func (m *PromMetrics) ObserveRequestDuration(duration time.Duration) {
	latencyHist.Observe(duration.Seconds())
}

func (m *PromMetrics) IncPVZCreated() {
	pvzCounter.Inc()
}

func (m *PromMetrics) IncReceptionCreated() {
	receptionCounter.Inc()
}

func (m *PromMetrics) IncProductAdded() {
	productCounter.Inc()
}
