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
	minQueryWaitTime    = 0
	minQueueSize        = 5
	defaultQueueSize    = 1000
	minDeleteWaitTime   = 1 * time.Millisecond
	minGetLimit         = 0
	defaultGetLimit     = 10
	minGetWaitTime      = 1 * time.Millisecond
)

var (
	defaultSleep  = time.Sleep
	defaultLogger = log.NewNopLogger()
)

type Config struct {
	Shard          int
	MaxWorkers     int
	MaxBatchSize   int
	QueueSize      int
	DeleteWaitTime time.Duration
	GetLimit       int
	GetWaitTime    time.Duration
}

type BatchDeleter struct {
	pruner        db.Pruner
	deleteQueue   chan []int
	deleteWorkers semaphore.Interface
	wg            sync.WaitGroup
	measures      *Measures
	logger        log.Logger
	config        Config
	sleep         func(time.Duration)
	stopTicker    func()
	stop          chan struct{}
}

func NewBatchDeleter(config Config, logger log.Logger, metricsRegistry provider.Provider, pruner db.Pruner) (*BatchDeleter, error) {
	if pruner == nil {
		return nil, errors.New("no pruner")
	}
	if config.MaxWorkers < minMaxWorkers {
		config.MaxWorkers = defaultMaxWorkers
	}
	if config.MaxBatchSize < minMaxBatchSize {
		config.MaxBatchSize = defaultMaxBatchSize
	}
	if config.QueueSize < minQueueSize {
		config.QueueSize = defaultQueueSize
	}
	if config.DeleteWaitTime < minDeleteWaitTime {
		config.DeleteWaitTime = minDeleteWaitTime
	}
	if config.GetLimit < minGetLimit {
		config.GetLimit = defaultGetLimit
	}
	if config.GetWaitTime < minGetWaitTime {
		config.GetWaitTime = minGetWaitTime
	}
	if logger == nil {
		logger = defaultLogger
	}

	measures := NewMeasures(metricsRegistry)
	workers := semaphore.New(config.MaxWorkers)
	queue := make(chan []int, config.QueueSize)
	stop := make(chan struct{}, 1)

	return &BatchDeleter{
		pruner:        pruner,
		deleteQueue:   queue,
		deleteWorkers: workers,
		config:        config,
		logger:        logger,
		sleep:         defaultSleep,
		stop:          stop,
		measures:      measures,
	}, nil
}

func (d *BatchDeleter) Start() {
	ticker := time.NewTicker(d.config.GetWaitTime)
	d.stopTicker = ticker.Stop
	d.wg.Add(2)
	go d.getRecordsToDelete(ticker.C)
	go d.delete()
}

func (d *BatchDeleter) Stop() {
	close(d.stop)
	d.wg.Wait()
}

func (d *BatchDeleter) getRecordsToDelete(ticker <-chan time.Time) {
	defer d.wg.Done()
	for {
		select {
		case <-d.stop:
			d.stopTicker()
			close(d.deleteQueue)
			return
		case <-ticker:
			vals, err := d.pruner.GetRecordIDs(d.config.Shard, d.config.GetLimit, time.Now().Unix())
			if err != nil {
				logging.Error(d.logger, emperror.Context(err)...).Log(logging.MessageKey(),
					"Failed to get record IDs from the database", logging.ErrorKey(), err.Error())
				// just in case
				vals = []int{}
			}
			logging.Debug(d.logger).Log(logging.MessageKey(), "got record ids", "record ids", vals)
			i := 0
			for i < len(vals) {
				endVal := i + d.config.MaxBatchSize
				if endVal > len(vals) {
					endVal = len(vals)
				}
				d.deleteQueue <- vals[i:endVal]
				if d.measures != nil {
					d.measures.DeletingQueue.Add(1.0)
				}
				i = endVal
			}
		}
	}
}

func (d *BatchDeleter) delete() {
	defer d.wg.Done()
	for records := range d.deleteQueue {
		if d.measures != nil {
			d.measures.DeletingQueue.Add(-1.0)
		}
		d.deleteWorkers.Acquire()
		go d.deleteWorker(records)
		d.sleep(d.config.DeleteWaitTime)
	}

	// Grab all the workers to make sure they are done.
	for i := 0; i < d.config.MaxWorkers; i++ {
		d.deleteWorkers.Acquire()
	}
}

func (d *BatchDeleter) deleteWorker(records []int) {
	defer d.deleteWorkers.Release()
	err := d.pruner.PruneRecords(records)
	if err != nil {
		logging.Error(d.logger, emperror.Context(err)...).Log(logging.MessageKey(),
			"Failed to delete records from the database", logging.ErrorKey(), err.Error())
		return
	}
	logging.Info(d.logger).Log(logging.MessageKey(), "Successfully deleted records", "records", records)
}
