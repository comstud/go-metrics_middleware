package metrics_middleware

import (
	"net/http"
	"time"
)

type MetricsManager interface {
	UpdateMetrics(*http.Request, *MetricsResponseWriter, time.Duration)
	GetMetrics() interface{}
}

// MetricsResponseWriter is used to trap status/size
type MetricsResponseWriter struct {
	http.ResponseWriter
	Status int
	Size   int
}

func (self *MetricsResponseWriter) Write(b []byte) (int, error) {
	size, err := self.ResponseWriter.Write(b)
	self.Size += size
	return size, err
}

func (self *MetricsResponseWriter) WriteHeader(s int) {
	self.ResponseWriter.WriteHeader(s)
	self.Status = s
}

type MetricsMiddleware struct {
	metricsManager MetricsManager
}

func (self *MetricsMiddleware) handler(h http.Handler, w http.ResponseWriter, r *http.Request) {
	ts := time.Now()
	w_wrap := &MetricsResponseWriter{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
	h.ServeHTTP(w_wrap, r)
	dur := time.Since(ts)
	self.metricsManager.UpdateMetrics(r, w_wrap, dur)
}

func (self *MetricsMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			self.handler(h, w, r)
		},
	)
}

func NewMiddleware(manager MetricsManager) *MetricsMiddleware {
	return &MetricsMiddleware{metricsManager: manager}
}

func NewHandler(manager MetricsManager, h http.Handler) http.Handler {
	mw := NewMiddleware(manager)
	return mw.Handler(h)
}
