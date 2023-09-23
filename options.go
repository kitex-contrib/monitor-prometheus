package prometheus

import (
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

// Option opts for monitor prometheus
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	buckets            []float64
	enableGoCollector  bool
	runtimeMetricRules []collectors.GoRuntimeMetricsRule
	registry           *prom.Registry
	serveMux           *http.ServeMux
	disableServer      bool
}

func defaultConfig() *config {
	return &config{
		buckets:            defaultBuckets,
		enableGoCollector:  false,
		runtimeMetricRules: []collectors.GoRuntimeMetricsRule{},
		registry:           prom.NewRegistry(),
		serveMux:           http.DefaultServeMux,
		disableServer:      false,
	}
}

// WithEnableGoCollector enable go collector
func WithEnableGoCollector(enable bool) Option {
	return option(func(cfg *config) {
		cfg.enableGoCollector = enable
	})
}

// WithGoCollectorRule define your custom go collector rule
func WithGoCollectorRule(rules ...collectors.GoRuntimeMetricsRule) Option {
	return option(func(cfg *config) {
		cfg.runtimeMetricRules = rules
	})
}

// WithHistogramBuckets define your custom histogram buckets base on your biz
func WithHistogramBuckets(buckets []float64) Option {
	return option(func(cfg *config) {
		if len(buckets) > 0 {
			cfg.buckets = buckets
		}
	})
}

// WithRegistry define your custom registry
func WithRegistry(registry *prom.Registry) Option {
	return option(func(cfg *config) {
		if registry != nil {
			cfg.registry = registry
		}
	})
}

// WithServeMux define your custom serve mux
func WithServeMux(serveMux *http.ServeMux) Option {
	return option(func(cfg *config) {
		if serveMux != nil {
			cfg.serveMux = serveMux
		}
	})
}

// WithDisableServer disable prometheus server
func WithDisableServer(disable bool) Option {
	return option(func(cfg *config) {
		cfg.disableServer = disable
	})
}
