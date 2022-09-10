/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package prometheus provides the extend implement of prometheus.
package prometheus

import (
	"log"
	"net/http"
	"sync"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloudwego/kitex/pkg/stats"
)

var gr globalRegistry

type globalRegistry struct {
	mtx                    sync.RWMutex
	registry               *prom.Registry
	addr                   string
	path                   string
	serverTracer           *serverTracer
	clientTracer           *clientTracer
	serverTracerRegistered bool
	clientTracerRegistered bool
	serverStarted          bool
	customizedRegistry     bool
}

// metricServerPrepare start the metric http server
func metricServerPrepare() {
	if gr.registry == nil {
		gr.registry = prom.NewRegistry()
		http.Handle(gr.path, promhttp.HandlerFor(gr.registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
		go func() {
			if !gr.serverStarted {
				if err := http.ListenAndServe(gr.addr, nil); err != nil {
					log.Fatal("Unable to start a promhttp server, err: " + err.Error())
				}
			}
		}()
	}
}

// NewGlobalClientTracer provide tracer for client call
func NewGlobalClientTracer() stats.Tracer {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()
	metricServerPrepare()

	// avoid register again
	if gr.clientTracerRegistered {
		return gr.clientTracer
	}

	clientHandledCounter := prom.NewCounterVec(
		prom.CounterOpts{
			Name: "kitex_client_throughput",
			Help: "Total number of RPCs completed by the client, regardless of success or failure.",
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	gr.registry.MustRegister(clientHandledCounter)

	clientHandledHistogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "kitex_client_latency_us",
			Help:    "Latency (microseconds) of the RPC until it is finished.",
			Buckets: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000},
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	gr.registry.MustRegister(clientHandledHistogram)
	gr.clientTracer = &clientTracer{
		clientHandledCounter:   clientHandledCounter,
		clientHandledHistogram: clientHandledHistogram,
	}
	gr.clientTracerRegistered = true

	return gr.clientTracer
}

// NewGlobalServerTracer provides tracer for server access
func NewGlobalServerTracer() stats.Tracer {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()
	metricServerPrepare()

	// avoid register again
	if gr.serverTracerRegistered {
		return gr.serverTracer
	}

	serverHandledCounter := prom.NewCounterVec(
		prom.CounterOpts{
			Name: "kitex_server_throughput",
			Help: "Total number of RPCs completed by the server, regardless of success or failure.",
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	gr.registry.MustRegister(serverHandledCounter)

	serverHandledHistogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "kitex_server_latency_us",
			Help:    "Latency (microseconds) of RPC that had been application-level handled by the server.",
			Buckets: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000},
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	gr.registry.MustRegister(serverHandledHistogram)
	gr.serverTracer = &serverTracer{
		serverHandledCounter:   serverHandledCounter,
		serverHandledHistogram: serverHandledHistogram,
	}
	gr.serverTracerRegistered = true

	return gr.serverTracer
}
