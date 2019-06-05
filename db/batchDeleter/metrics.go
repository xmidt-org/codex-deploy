/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
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
 *
 */

package batchDeleter

import (
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

const (
	DeletingQueueDepth = "deleting_queue_depth"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: DeletingQueueDepth,
			Help: "The depth of the delete queue",
			Type: "gauge",
		},
	}
}

type Measures struct {
	DeletingQueue metrics.Gauge
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		DeletingQueue: p.NewGauge(DeletingQueueDepth),
	}
}
