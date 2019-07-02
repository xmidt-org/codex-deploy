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
	"errors"
	"github.com/Comcast/codex/capacityset"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Comcast/webpa-common/semaphore"

	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/xmetrics/xmetricstest"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/stretchr/testify/assert"
)

func TestNewBatchDeleter(t *testing.T) {
	goodPruner := new(mockPruner)
	goodRegistry := xmetricstest.NewProvider(nil, Metrics)
	goodMeasures := NewMeasures(goodRegistry)
	goodConfig := Config{
		SetSize:        1000,
		MaxWorkers:     5000,
		MaxBatchSize:   100,
		DeleteWaitTime: 5 * time.Hour,
		GetLimit:       1000000,
		GetWaitTime:    2 * time.Hour,
	}
	tests := []struct {
		description          string
		config               Config
		pruner               db.Pruner
		logger               log.Logger
		registry             provider.Provider
		expectedBatchDeleter *BatchDeleter
		expectedErr          error
	}{
		{
			description: "Success",
			config:      goodConfig,
			pruner:      goodPruner,
			logger:      log.NewJSONLogger(os.Stdout),
			registry:    goodRegistry,
			expectedBatchDeleter: &BatchDeleter{
				pruner:   goodPruner,
				measures: goodMeasures,
				config:   goodConfig,
				logger:   log.NewJSONLogger(os.Stdout),
			},
		},
		{
			description: "Success With Defaults",
			config: Config{
				MaxBatchSize:   -5,
				DeleteWaitTime: -2 * time.Minute,
				GetLimit:       -3,
				GetWaitTime:    -2 * time.Minute,
			},
			pruner:   goodPruner,
			registry: goodRegistry,
			expectedBatchDeleter: &BatchDeleter{
				pruner:   goodPruner,
				measures: goodMeasures,
				config: Config{
					MaxBatchSize:   defaultMaxBatchSize,
					DeleteWaitTime: minDeleteWaitTime,
					SetSize:        defaultSetSize,
					MaxWorkers:     defaultMaxWorkers,
					GetLimit:       defaultGetLimit,
					GetWaitTime:    minDeleteWaitTime,
				},
				logger: defaultLogger,
			},
		},
		{
			description: "Nil Pruner Error",
			expectedErr: errors.New("no pruner"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			bd, err := NewBatchDeleter(tc.config, tc.logger, tc.registry, tc.pruner)
			if bd != nil {
			}
			if tc.expectedBatchDeleter == nil || bd == nil {
				assert.Equal(tc.expectedBatchDeleter, bd)
			} else {
				assert.Equal(tc.expectedBatchDeleter.pruner, bd.pruner)
				assert.Equal(tc.expectedBatchDeleter.measures, bd.measures)
				assert.Equal(tc.expectedBatchDeleter.config, bd.config)
				assert.Equal(tc.expectedBatchDeleter.logger, bd.logger)
			}
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetRecordsToDeleteSuccess(t *testing.T) {
	assert := assert.New(t)
	vals := []db.RecordToDelete{{DeathDate: 1, RecordID: 2}, {DeathDate: 3, RecordID: 4}}
	pruner := new(mockPruner)
	pruner.On("GetRecordsToDelete", mock.Anything, mock.Anything, mock.Anything).Return(vals, nil).Once()
	tickerChan := make(chan time.Time, 1)
	stopChan := make(chan struct{}, 1)
	p := xmetricstest.NewProvider(nil, Metrics)
	measures := NewMeasures(p)

	stopCalled := false
	stopFunc := func() {
		stopCalled = true
	}

	batchDeleter := &BatchDeleter{
		pruner:        pruner,
		logger:        defaultLogger,
		deleteSet:     capacityset.NewCapacitySet(2),
		deleteWorkers: semaphore.New(3),
		measures:      measures,
		config: Config{
			MaxBatchSize: 3,
		},
		stop:       stopChan,
		stopTicker: stopFunc,
		deleteStop: make(chan struct{}, 1),
	}

	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(0))
	batchDeleter.wg.Add(1)
	tickerChan <- time.Now()
	go batchDeleter.getRecordsToDelete(tickerChan)
	time.Sleep(1 * time.Second)
	batchDeleter.Stop()

	pruner.AssertExpectations(t)
	assert.True(stopCalled)
	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(2))
}

func TestGetRecordsToDeleteError(t *testing.T) {
	assert := assert.New(t)
	pruner := new(mockPruner)
	pruner.On("GetRecordsToDelete", mock.Anything, mock.Anything, mock.Anything).Return([]db.RecordToDelete{}, errors.New("test error")).Once()
	tickerChan := make(chan time.Time, 1)
	stopChan := make(chan struct{}, 1)
	p := xmetricstest.NewProvider(nil, Metrics)
	measures := NewMeasures(p)

	stopCalled := false
	stopFunc := func() {
		stopCalled = true
	}

	batchDeleter := &BatchDeleter{
		pruner:        pruner,
		logger:        defaultLogger,
		deleteSet:     capacityset.NewCapacitySet(2),
		deleteWorkers: semaphore.New(3),
		measures:      measures,
		config: Config{
			MaxBatchSize: 3,
		},
		stop:       stopChan,
		stopTicker: stopFunc,
		deleteStop: make(chan struct{}, 1),
	}

	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(0))
	batchDeleter.wg.Add(1)
	tickerChan <- time.Now()
	go batchDeleter.getRecordsToDelete(tickerChan)
	time.Sleep(1 * time.Second)
	batchDeleter.Stop()

	pruner.AssertExpectations(t)
	assert.True(stopCalled)
	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(0))
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	vals := db.RecordToDelete{DeathDate: 111, RecordID: 88888}
	pruner := new(mockPruner)
	pruner.On("DeleteRecord", 0, vals.DeathDate, vals.RecordID).Return(nil).Once()
	p := xmetricstest.NewProvider(nil, Metrics)
	measures := NewMeasures(p)

	sleepCalled := false
	sleepTime := 5 * time.Minute
	sleepFunc := func(t time.Duration) {
		sleepCalled = true
		assert.Equal(sleepTime, t)
	}

	batchDeleter := &BatchDeleter{
		pruner:        pruner,
		logger:        defaultLogger,
		deleteSet:     capacityset.NewCapacitySet(2),
		deleteWorkers: semaphore.New(3),
		measures:      measures,
		config: Config{
			MaxWorkers:     3,
			DeleteWaitTime: sleepTime,
		},
		sleep:      sleepFunc,
		deleteStop: make(chan struct{}, 1),
	}

	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(0))
	batchDeleter.wg.Add(1)
	batchDeleter.deleteSet.Add(vals)
	go batchDeleter.delete()
	time.Sleep(time.Second)
	batchDeleter.deleteStop <- struct{}{}
	batchDeleter.wg.Wait()

	pruner.AssertExpectations(t)
	assert.True(sleepCalled)
	p.Assert(t, DeletingQueueDepth)(xmetricstest.Value(-1))
}
