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
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prometheus

import (
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
)

// counterAdd wraps Add of prom.Counter.
func counterAdd(counterVec *prom.CounterVec, value int, labels prom.Labels) error {
	counter, err := counterVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

// histogramObserve wraps Observe of prom.Observer.
func histogramObserve(histogramVec *prom.HistogramVec, value time.Duration, labels prom.Labels) error {
	histogram, err := histogramVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	histogram.Observe(float64(value.Microseconds()))
	return nil
}
