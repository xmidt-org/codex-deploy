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

package db

import (
	"errors"
	"testing"
	"time"

	"github.com/Comcast/webpa-common/xmetrics/xmetricstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryInsertRecords(t *testing.T) {
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
					retries:  tc.retries,
					interval: interval,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			err := retryInsertService.InsertRecords(Record{})
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, typeLabel, insertType)(xmetricstest.Value(tc.expectedRetryMetric))
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
			retries:  322,
			interval: 2 * time.Minute,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryInsertService(r.inserter, r.config.retries, r.config.interval, p)
	assert.Equal(r.inserter, newService.inserter)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
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
				mockObj.On("PruneRecords", mock.Anything, mock.Anything, mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("PruneRecords", mock.Anything, mock.Anything, mock.Anything).Return(tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryInsertService := RetryUpdateService{
				pruner: mockObj,
				config: retryConfig{
					retries:  tc.retries,
					interval: interval,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			err := retryInsertService.PruneRecords(time.Now().Unix())
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, typeLabel, deleteType)(xmetricstest.Value(tc.expectedRetryMetric))
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
			retries:  322,
			interval: 2 * time.Minute,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryUpdateService(r.pruner, r.config.retries, r.config.interval, p)
	assert.Equal(r.pruner, newService.pruner)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
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
				mockObj.On("GetRecords", mock.Anything).Return([]Record{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetRecords", mock.Anything).Return([]Record{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryRGService := RetryRGService{
				rg: mockObj,
				config: retryConfig{
					retries:  tc.retries,
					interval: interval,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			_, err := retryRGService.GetRecords("")
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, typeLabel, readType)(xmetricstest.Value(tc.expectedRetryMetric))
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
				mockObj.On("GetRecordsOfType", mock.Anything, mock.Anything).Return([]Record{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetRecordsOfType", mock.Anything, mock.Anything).Return([]Record{}, tc.finalError).Once()
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryRGService := RetryRGService{
				rg: mockObj,
				config: retryConfig{
					retries:  tc.retries,
					interval: interval,
					sleep: func(t time.Duration) {
						assert.Equal(interval, t)
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			_, err := retryRGService.GetRecordsOfType("", 0)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQueryRetryCounter, typeLabel, readType)(xmetricstest.Value(tc.expectedRetryMetric))
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
			retries:  322,
			interval: 2 * time.Minute,
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryRGService(r.rg, r.config.retries, r.config.interval, p)
	assert.Equal(r.rg, newService.rg)
	assert.Equal(r.config.retries, newService.config.retries)
	assert.Equal(r.config.interval, newService.config.interval)
}
