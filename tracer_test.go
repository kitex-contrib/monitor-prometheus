/*
 * Copyright 2021 CloudWeGo Authors
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
 * See the License for the specific language governing permissions a nd
 * limitations under the License.
 */

package prometheus

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusReporter(t *testing.T) {
	registry := prom.NewRegistry()
	http.Handle("/prometheus", promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	go func() {
		if err := http.ListenAndServe(":9090", nil); err != nil {
			klog.Fatal("Unable to start a promhttp server, err: " + err.Error())
		}
	}()

	counter := prom.NewCounterVec(
		prom.CounterOpts{
			Name:        "test_counter",
			ConstLabels: prom.Labels{"service": "prometheus-test"},
		},
		[]string{"test1", "test2"},
	)
	registry.MustRegister(counter)

	histogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:        "test_histogram",
			ConstLabels: prom.Labels{"service": "prometheus-test"},
			Buckets:     prom.DefBuckets,
		},
		[]string{"test1", "test2"},
	)
	registry.MustRegister(histogram)

	labels := prom.Labels{
		"test1": "abc",
		"test2": "def",
	}

	assert.True(t, counterAdd(counter, 6, labels) == nil)
	assert.True(t, histogramObserve(histogram, time.Second, labels) == nil)

	time.Sleep(time.Second) // wait server start

	promServerResp, err := http.Get("http://localhost:9090/prometheus")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, promServerResp.StatusCode == http.StatusOK)

	bodyBytes, err := ioutil.ReadAll(promServerResp.Body)
	assert.True(t, err == nil)
	respStr := string(bodyBytes)
	assert.True(t, strings.Contains(respStr, `test_counter{service="prometheus-test",test1="abc",test2="def"} 6`))
	assert.True(t, strings.Contains(respStr, `test_histogram_sum{service="prometheus-test",test1="abc",test2="def"} 1e+06`))
}
