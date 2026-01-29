package middleware

import (
	"net/http"
	"time"

	"pvz-service/internal/usecase/ports"
)

func MetricsMiddleware(metrics ports.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			metrics.IncRequest()
			metrics.ObserveRequestDuration(duration)
		})
	}
}
