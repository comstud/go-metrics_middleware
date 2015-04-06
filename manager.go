package metrics_middleware

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"time"
)

const BYTES = "Bytes"

type RouteManager interface {
	RouteInfoForRequest(*http.Request) (method, path string)
}

type DefaultMetricsManager struct {
	Registry     *metrics.StandardRegistry
	routeManager RouteManager
}

func (self *DefaultMetricsManager) updateTimers(dur time.Duration, names ...string) {
	for _, name := range names {
		// Always doing a Get first saves an allocation when it exists.
		t := self.Registry.Get(name)
		if t == nil {
			t = metrics.NewTimer()
			t = self.Registry.GetOrRegister(name, t)
		}
		t.(metrics.Timer).Update(dur)
	}
}

func (self *DefaultMetricsManager) updateCounts(cnt int64, names ...string) {
	for _, name := range names {
		// Always doing a Get first saves an allocation when it exists.
		c := self.Registry.Get(name)
		if c == nil {
			c = metrics.NewCounter()
			c = self.Registry.GetOrRegister(name, c)
		}
		c.(metrics.Counter).Inc(cnt)
	}
}

func (self *DefaultMetricsManager) AddRoute(method, path string) {
	key := path + ":" + method
	t := metrics.NewTimer()
	self.Registry.Register(key, t)
	key += ":" + BYTES
	c := metrics.NewCounter()
	self.Registry.Register(key, c)
}

func (self *DefaultMetricsManager) GetMetrics() interface{} {
	return self.Registry
}

func (self *DefaultMetricsManager) UpdateMetrics(r *http.Request, w *MetricsResponseWriter, dur time.Duration) {
	method, path := self.routeManager.RouteInfoForRequest(r)
	status := w.Status
	size := w.Size
	go func() {
		key := path + ":" + method
		self.updateTimers(
			dur,
			method,
			fmt.Sprintf("%s:%d", method, status),
			key,
			fmt.Sprintf("%s:%d", key, status),
		)
		self.updateCounts(
			int64(size),
			method+":"+BYTES,
			key+":"+BYTES,
		)
	}()
}

type DefaultRouteManager struct{}

func (self *DefaultRouteManager) RouteInfoForRequest(r *http.Request) (method, path string) {
	return r.Method, r.URL.Path
}

func NewCustomMetricsManager(r RouteManager) *DefaultMetricsManager {
	return &DefaultMetricsManager{
		Registry:     metrics.NewRegistry().(*metrics.StandardRegistry),
		routeManager: r,
	}
}

func NewDefaultMetricsManager() *DefaultMetricsManager {
	return NewCustomMetricsManager(&DefaultRouteManager{})
}
