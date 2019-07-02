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

// package batchDeleter provides a wrapper around the db.Pruner to provide a
// way to get expired records at a given interval and delete them at a separate
// given interval.
package batchDeleter

import (
	"errors"
	"github.com/Comcast/codex/capacityset"
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
	minMaxBatchSize     = 1
	defaultMaxBatchSize = 1
	minQueryWaitTime    = 0
	minSetSize          = 5
	defaultSetSize      = 1000
	minDeleteWaitTime   = 1 * time.Millisecond
	minGetLimit         = 0
	defaultGetLimit     = 10
	minGetWaitTime      = 1 * time.Millisecond
)

var (
	defaultSleep  = time.Sleep
	defaultLogger = log.NewNopLogger()
)

// Config holds the configuration values for a batch deleter.
type Config struct {
	Shard          int
	MaxWorkers     int
	MaxBatchSize   int
	SetSize        int
	DeleteWaitTime time.Duration
	GetLimit       int
	GetWaitTime    time.Duration
}

// BatchDeleter manages getting records that have expired and then deleting
// them.
type BatchDeleter struct {
	pruner        db.Pruner
	deleteSet     capacityset.Set
	deleteWorkers semaphore.Interface
	wg            sync.WaitGroup
	measures      *Measures
	logger        log.Logger
	config        Config
	sleep         func(time.Duration)
	stopTicker    func()
	stop          chan struct{}
	deleteStop    chan struct{}
}

// NewBatchDeleter creates a BatchDeleter with the given values, ensuring
// that the configuration and other values given are valid.  If configuration
// values aren't valid, a default value is used.
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
	if config.SetSize < minSetSize {
		config.SetSize = defaultSetSize
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
	stop := make(chan struct{}, 1)

	return &BatchDeleter{
		pruner:        pruner,
		deleteSet:     capacityset.NewCapacitySet(config.SetSize),
		deleteWorkers: workers,
		config:        config,
		logger:        logger,
		sleep:         defaultSleep,
		stop:          stop,
		measures:      measures,
		deleteStop:    make(chan struct{}, 1),
	}, nil
}

// Start starts the batcher, which includes a ticker for getting expired
// records at an interval and the workers that do the deleting.
func (d *BatchDeleter) Start() {
	ticker := time.NewTicker(d.config.GetWaitTime)
	d.stopTicker = ticker.Stop
	d.wg.Add(2)
	go d.getRecordsToDelete(ticker.C)
	go d.delete()
}

// Stop closes the internal queue and waits for the workers to finish
// processing what has already been added.  This can block as it waits for
// everything to stop.
func (d *BatchDeleter) Stop() {
	close(d.stop)
	d.deleteStop <- struct{}{}
	d.wg.Wait()
}

func (d *BatchDeleter) getRecordsToDelete(ticker <-chan time.Time) {
	defer d.wg.Done()
	for {
		select {
		case <-d.stop:
			d.stopTicker()
			return
		case <-ticker:
			vals, err := d.pruner.GetRecordsToDelete(d.config.Shard, d.config.GetLimit, time.Now().UnixNano())
			if err != nil {
				logging.Error(d.logger, emperror.Context(err)...).Log(logging.MessageKey(),
					"Failed to get record IDs from the database", logging.ErrorKey(), err.Error())
				// just in case
				// vals = []int{}
			}
			logging.Debug(d.logger).Log(logging.MessageKey(), "got records", "records", vals)
			// i := 0
			// for i < len(vals) {
			// 	endVal := i + d.config.MaxBatchSize
			// 	if endVal > len(vals) {
			// 		endVal = len(vals)
			// 	}
			// 	d.deleteQueue <- vals[i:endVal]
			// 	if d.measures != nil {
			// 		d.measures.DeletingQueue.Add(1.0)
			// 	}
			// 	i = endVal
			// }

			for _, i := range vals {
				if d.deleteSet.Add(i) {
					if d.measures != nil {
						d.measures.DeletingQueue.Add(1.0)
					}
				}
			}
		}
	}
}

func (d *BatchDeleter) delete() {
	defer d.wg.Done()

deleteLoop:
	for {
		select {
		case <-d.deleteStop:
			break deleteLoop
		case item := <-capacityset.WrapBlockingCall(d.deleteSet.Pop):
			if item == nil {
				continue
			}
			record := item.(db.RecordToDelete)
			if d.measures != nil {
				d.measures.DeletingQueue.Add(-1.0)
			}
			d.deleteWorkers.Acquire()
			go d.deleteWorker(record)
			d.sleep(d.config.DeleteWaitTime)
		}
	}

	// Grab all the workers to make sure they are done.
	for i := 0; i < d.config.MaxWorkers; i++ {
		d.deleteWorkers.Acquire()
	}
}

func (d *BatchDeleter) deleteWorker(record db.RecordToDelete) {
	defer d.deleteWorkers.Release()
	err := d.pruner.DeleteRecord(d.config.Shard, record.DeathDate, record.RecordID)
	if err != nil {
		logging.Error(d.logger, emperror.Context(err)...).Log(logging.MessageKey(),
			"Failed to delete records from the database", logging.ErrorKey(), err.Error())
		return
	}
	logging.Debug(d.logger).Log(logging.MessageKey(), "Successfully deleted record", "record id", record.RecordID)
}
