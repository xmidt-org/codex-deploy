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

package batchInserter

import (
	"errors"
	"sync"
	"time"

	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/logging"
	"github.com/Comcast/webpa-common/semaphore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"
)

const (
	minMaxWorkers       = 1
	defaultMaxWorkers   = 5
	minMaxBatchSize     = 0
	defaultMaxBatchSize = 1
	minMaxBatchWaitTime = time.Duration(1) * time.Millisecond
	defaultMinQueueSize = 5
)

var (
	defaultLogger = log.NewNopLogger()
)

// defaultTicker is the production code that produces a ticker.  Note that we don't
// want to return *time.Ticker, as we want to be able to inject something for testing.
// We also need to return a closure to stop the ticker, so that we can call ticker.Stop() without
// being dependent on the *time.Ticker interface.
func defaultTicker(d time.Duration) (<-chan time.Time, func()) {
	ticker := time.NewTicker(d)
	return ticker.C, ticker.Stop
}

type BatchInserter struct {
	insertQueue   chan db.Record
	inserter      db.Inserter
	insertWorkers semaphore.Interface
	wg            sync.WaitGroup
	measures      *Measures
	logger        log.Logger
	config        Config
	ticker        func(time.Duration) (<-chan time.Time, func())
}

type Config struct {
	MaxWorkers       int
	MaxBatchSize     int
	MaxBatchWaitTime time.Duration
	QueueSize        int
}

func NewBatchInserter(config Config, logger log.Logger, metricsRegistry provider.Provider, inserter db.Inserter) (*BatchInserter, error) {
	if inserter == nil {
		return nil, errors.New("no inserter")
	}
	if config.MaxWorkers < minMaxWorkers {
		config.MaxWorkers = defaultMaxWorkers
	}
	if config.MaxBatchSize < minMaxBatchSize {
		config.MaxBatchSize = defaultMaxBatchSize
	}
	if config.MaxBatchWaitTime < minMaxBatchWaitTime {
		config.MaxBatchWaitTime = minMaxBatchWaitTime
	}
	if config.QueueSize < defaultMinQueueSize {
		config.QueueSize = defaultMinQueueSize
	}
	if logger == nil {
		logger = defaultLogger
	}

	measures := NewMeasures(metricsRegistry)
	workers := semaphore.New(config.MaxWorkers)
	queue := make(chan db.Record, config.QueueSize)
	b := BatchInserter{
		config:        config,
		logger:        logger,
		measures:      measures,
		insertWorkers: workers,
		inserter:      inserter,
		insertQueue:   queue,
		ticker:        defaultTicker,
	}
	return &b, nil
}

func (b *BatchInserter) Start() {
	b.wg.Add(1)
	go b.batchRecords()
}

func (b *BatchInserter) Insert(record db.Record) {
	b.insertQueue <- record
	if b.measures != nil {
		b.measures.InsertingQueue.Add(1.0)
	}
}

func (b *BatchInserter) Stop() {
	close(b.insertQueue)
	b.wg.Wait()
}

func (b *BatchInserter) batchRecords() {
	var (
		insertRecords bool
		ticker        <-chan time.Time
		stop          func()
	)
	defer b.wg.Done()
	for record := range b.insertQueue {
		if b.measures != nil {
			b.measures.InsertingQueue.Add(-1.0)
		}
		if record.Data == nil || len(record.Data) == 0 {
			continue
		}
		ticker, stop = b.ticker(b.config.MaxBatchWaitTime)
		records := []db.Record{record}
		for {
			select {
			case <-ticker:
				insertRecords = true
			case r := <-b.insertQueue:
				if b.measures != nil {
					b.measures.InsertingQueue.Add(-1.0)
				}
				if r.Data == nil || len(r.Data) == 0 {
					continue
				}
				records = append(records, r)
				if b.config.MaxBatchSize != 0 && len(records) >= b.config.MaxBatchSize {
					insertRecords = true
				}
			}
			if insertRecords {
				b.insertWorkers.Acquire()
				go b.insertRecords(records)
				insertRecords = false
				break
			}
		}
		stop()
	}

	// Grab all the workers to make sure they are done.
	for i := 0; i < b.config.MaxWorkers; i++ {
		b.insertWorkers.Acquire()
	}
}

func (b *BatchInserter) insertRecords(records []db.Record) {
	defer b.insertWorkers.Release()
	err := b.inserter.InsertRecords(records...)
	if err != nil {
		if b.measures != nil {
			b.measures.DroppedEventsFromDbFailCount.Add(float64(len(records)))
		}
		logging.Error(b.logger, emperror.Context(err)...).Log(logging.MessageKey(),
			"Failed to add records to the database", logging.ErrorKey(), err.Error())
		return
	}
	logging.Debug(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", records)
	logging.Info(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", len(records))
}
