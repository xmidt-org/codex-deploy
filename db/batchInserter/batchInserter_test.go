package batchInserter

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Comcast/webpa-common/semaphore"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/stretchr/testify/assert"

	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/xmetrics/xmetricstest"
)

func TestNewBatchInserter(t *testing.T) {
	goodInserter := new(mockInserter)
	goodRegistry := xmetricstest.NewProvider(nil, Metrics)
	goodMeasures := NewMeasures(goodRegistry)
	goodConfig := Config{
		QueueSize:        1000,
		MaxWorkers:       5000,
		MaxBatchSize:     100,
		MaxBatchWaitTime: 5 * time.Hour,
	}
	tests := []struct {
		description           string
		config                Config
		inserter              db.Inserter
		logger                log.Logger
		registry              provider.Provider
		expectedBatchInserter *BatchInserter
		expectedErr           error
	}{
		{
			description: "Success",
			config:      goodConfig,
			inserter:    goodInserter,
			logger:      log.NewJSONLogger(os.Stdout),
			registry:    goodRegistry,
			expectedBatchInserter: &BatchInserter{
				inserter: goodInserter,
				measures: goodMeasures,
				config:   goodConfig,
				logger:   log.NewJSONLogger(os.Stdout),
			},
		},
		{
			description: "Success With Defaults",
			config: Config{
				MaxBatchSize:     -5,
				MaxBatchWaitTime: -2 * time.Minute,
			},
			inserter: goodInserter,
			registry: goodRegistry,
			expectedBatchInserter: &BatchInserter{
				inserter: goodInserter,
				measures: goodMeasures,
				config: Config{
					MaxBatchSize:     defaultMaxBatchSize,
					MaxBatchWaitTime: minMaxBatchWaitTime,
					QueueSize:        defaultMinQueueSize,
					MaxWorkers:       defaultMaxWorkers,
				},
				logger: defaultLogger,
			},
		},
		{
			description: "Nil Inserter Error",
			expectedErr: errors.New("no inserter"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			bi, err := NewBatchInserter(tc.config, tc.logger, tc.registry, tc.inserter)
			if bi != nil {
				tc.expectedBatchInserter.insertQueue = bi.insertQueue
				tc.expectedBatchInserter.insertWorkers = bi.insertWorkers
			}
			assert.Equal(tc.expectedBatchInserter, bi)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestBatchInserter(t *testing.T) {
	records := []db.Record{
		{
			Type: db.State,
			Data: []byte("test1"),
		},
		{
			Type: db.State,
			Data: []byte("test2"),
		},
		{
			Type: db.State,
			Data: []byte("test3"),
		},
		{
			Type: db.State,
			Data: []byte("test4"),
		},
		{
			Type: db.State,
			Data: []byte("test5"),
		},
	}
	tests := []struct {
		description           string
		insertErr             error
		recordsToInsert       []db.Record
		recordsExpected       [][]db.Record
		waitBtwnRecords       time.Duration
		expectedDroppedEvents float64
	}{
		{
			description:     "Success",
			waitBtwnRecords: 1 * time.Millisecond,
			recordsToInsert: records[:5],
			recordsExpected: [][]db.Record{
				records[:3],
				records[3:5],
			},
		},
		{
			description:     "Nil Record",
			recordsToInsert: []db.Record{{}},
		},
		{
			description:     "Insert Records Error",
			recordsToInsert: records[3:5],
			waitBtwnRecords: 1 * time.Millisecond,
			recordsExpected: [][]db.Record{
				records[3:5],
			},
			insertErr:             errors.New("test insert error"),
			expectedDroppedEvents: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			inserter := new(mockInserter)
			for _, r := range tc.recordsExpected {
				inserter.On("InsertRecords", r).Return(tc.insertErr).Once()
			}
			queue := make(chan db.Record, 5)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			b := BatchInserter{
				config: Config{
					MaxBatchWaitTime: 10 * time.Millisecond,
					MaxBatchSize:     3,
					MaxWorkers:       5,
				},
				inserter:      inserter,
				insertQueue:   queue,
				insertWorkers: semaphore.New(5),
				measures:      m,
				logger:        log.NewNopLogger(),
			}
			p.Assert(t, DroppedEventsFromDbFailCounter)(xmetricstest.Value(0))
			b.wg.Add(1)
			go b.batchRecords()
			for i, r := range tc.recordsToInsert {
				if i > 0 {
					time.Sleep(tc.waitBtwnRecords)
				}
				b.Insert(r)
			}
			b.Stop()
			inserter.AssertExpectations(t)
			p.Assert(t, DroppedEventsFromDbFailCounter)(xmetricstest.Value(tc.expectedDroppedEvents))
		})
	}
}
