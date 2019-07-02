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
	"errors"
	"testing"
	"time"

	"github.com/Comcast/codex/blacklist"
	"github.com/Comcast/codex/db"

	"github.com/Comcast/webpa-common/xmetrics/xmetricstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryInsertRecords(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := time.Duration(8) * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockInserter)
			if tc.numCalls > 1 {
				mockObj.On("InsertRecords", mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("InsertRecords", mock.Anything).Return(tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryInsertService := RetryInsertService{
				inserter: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			err := retryInsertService.InsertRecords(db.Record{})
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestCreateRetryInsertService(t *testing.T) {
	r := RetryInsertService{
		inserter: new(mockInserter),
		config: retryConfig{
			retries:      322,
			interval:     2 * time.Minute,
			intervalMult: 1,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryInsertService(r.inserter, WithRetries(r.config.retries), WithInterval(r.config.interval), WithMeasures(p))
	assert.Equal(r.inserter, newService.inserter)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
	assert.Equal(r.config.intervalMult, newService.config.intervalMult)
}

func TestRetryGetRecordIDs(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockPruner)
			if tc.numCalls > 1 {
				mockObj.On("GetRecordsToDelete", mock.Anything, mock.Anything, mock.Anything).Return([]db.RecordToDelete{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetRecordsToDelete", mock.Anything, mock.Anything, mock.Anything).Return([]db.RecordToDelete{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryUpdateService := RetryUpdateService{
				pruner: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			_, err := retryUpdateService.GetRecordsToDelete(0, 0, time.Now().UnixNano())
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestRetryPruneRecords(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockPruner)
			if tc.numCalls > 1 {
				mockObj.On("DeleteRecord", mock.Anything, mock.Anything, mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("DeleteRecord", mock.Anything, mock.Anything, mock.Anything).Return(tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryUpdateService := RetryUpdateService{
				pruner: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			err := retryUpdateService.DeleteRecord(0, 0, 0)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.DeleteType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.DeleteType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryUpdateService(t *testing.T) {
	r := RetryUpdateService{
		pruner: new(mockPruner),
		config: retryConfig{
			retries:      322,
			interval:     2 * time.Minute,
			intervalMult: 8,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryUpdateService(r.pruner, WithRetries(r.config.retries), WithInterval(r.config.interval), WithIntervalMultiplier(8), WithMeasures(p))
	assert.Equal(r.pruner, newService.pruner)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
	assert.Equal(r.config.intervalMult, newService.config.intervalMult)
}

func TestRetryGetBlacklist(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockLG)
			if tc.numCalls > 1 {
				mockObj.On("GetBlacklist").Return([]blacklist.BlackListedItem{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetBlacklist").Return([]blacklist.BlackListedItem{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryListGService := RetryListGService{
				lg: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			_, err := retryListGService.GetBlacklist()
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.BlacklistReadType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.BlacklistReadType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryListGService(t *testing.T) {
	r := RetryListGService{
		lg: new(mockLG),
		config: retryConfig{
			retries:      322,
			interval:     2 * time.Minute,
			intervalMult: 5,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryListGService(r.lg, WithRetries(r.config.retries), WithInterval(r.config.interval), WithIntervalMultiplier(5), WithMeasures(p))
	assert.Equal(r.lg, newService.lg)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
	assert.Equal(r.config.intervalMult, newService.config.intervalMult)
}

func TestRetryGetRecords(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockRG)
			if tc.numCalls > 1 {
				mockObj.On("GetRecords", mock.Anything, mock.Anything).Return([]db.Record{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetRecords", mock.Anything, mock.Anything).Return([]db.Record{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryRGService := RetryRGService{
				rg: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			_, err := retryRGService.GetRecords("", 5)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestRetryGetRecordsOfType(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description         string
		numCalls            int
		retries             int
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			retries:             5,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			retries:             3,
			expectedRetryMetric: 3.0,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockRG)
			if tc.numCalls > 1 {
				mockObj.On("GetRecordsOfType", mock.Anything, mock.Anything, mock.Anything).Return([]db.Record{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetRecordsOfType", mock.Anything, mock.Anything, mock.Anything).Return([]db.Record{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryRGService := RetryRGService{
				rg: mockObj,
				config: retryConfig{
					retries:      tc.retries,
					interval:     interval,
					intervalMult: 1,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			_, err := retryRGService.GetRecordsOfType("", 5, 0)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedRetryMetric))
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryRGService(t *testing.T) {
	r := RetryRGService{
		rg: new(mockRG),
		config: retryConfig{
			retries:      322,
			interval:     2 * time.Minute,
			intervalMult: 5,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryRGService(r.rg, WithRetries(r.config.retries), WithInterval(r.config.interval), WithIntervalMultiplier(5), WithMeasures(p))
	assert.Equal(r.rg, newService.rg)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
	assert.Equal(r.config.intervalMult, newService.config.intervalMult)
}
