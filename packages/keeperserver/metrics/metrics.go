package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_events_total",
			Help: "Total number of events to HTTP",
		}, []string{"method", "path", "status_code"},
	)
)

func init(){
	prometheus.MustRegister(HTTPEventsTotal)
}