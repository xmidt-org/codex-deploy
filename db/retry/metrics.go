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

package dbretry

import (
	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

const (
	SQLQueryRetryCounter = "sql_query_retry_count"
	SQLQueryEndCounter   = "sql_query_end_counter"
)

//Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name:       SQLQueryRetryCounter,
			Type:       "counter",
			Help:       "The total number of SQL queries retried",
			LabelNames: []string{db.TypeLabel},
		},
		{
			Name:       SQLQueryEndCounter,
			Type:       "counter",
			Help:       "the total number of SQL queries that are done, no more retrying",
			LabelNames: []string{db.TypeLabel},
		},
	}
}

type Measures struct {
	SQLQueryRetryCount metrics.Counter
	SQLQueryEndCount   metrics.Counter
}

func NewMeasures(p provider.Provider) Measures {
	return Measures{
		SQLQueryRetryCount: p.NewCounter(SQLQueryRetryCounter),
		SQLQueryEndCount:   p.NewCounter(SQLQueryEndCounter),
	}
}
