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

// package dbretry contains structs that implement various db interfaces as
// well as consume them.  They allow consumers to easily try to interact with
// the database a configurable number of times, with configurable backoff
// options and metrics.
package dbretry

import (
	"time"

	"github.com/Comcast/codex/blacklist"
	"github.com/Comcast/codex/db"

	"github.com/go-kit/kit/metrics/provider"
)

const (
	defaultInterval     = time.Second
	defaultIntervalMult = 1
	defaultRetries      = 1
)

var (
	defaultSleep = time.Sleep
)

type retryConfig struct {
	retries      int
	interval     time.Duration
	intervalMult time.Duration
	sleep        func(time.Duration)
	measures     Measures
}

// Option is the function used to configure the retry objects.
type Option func(r *retryConfig)

// WithRetries sets the number of times to potentially try to interact with the
// database if the inital attempt doesn't succeed.
func WithRetries(retries int) Option {
	return func(r *retryConfig) {
		// only set retries if the value is valid
		if retries >= 0 {
			r.retries = retries
		}
	}
}

// WithInterval sets the amount of time to wait between the initial attempt and
// the first retry.  If the interval multiplier is 1, this interval is used
// between every attempt.
func WithInterval(interval time.Duration) Option {
	return func(r *retryConfig) {
		// only set interval if the value is valid
		if interval > time.Duration(0)*time.Second {
			r.interval = interval
		}
	}
}

// WithIntervalMultiplier sets the interval multiplier, which is multiplied
// against the interval time for each wait time after the first retry.  For
// example, if the interval is 1s, the interval multiplier 5, and the number of
// retries 3, then between the initial attempt and first retry, the program
// will wait 1s.  Between the first retry and the second retry, the program
// will wait 5s.  Between the second retry and the third, the program will
// wait 25s.  This is assuming all attempts fail.
func WithIntervalMultiplier(mult time.Duration) Option {
	return func(r *retryConfig) {
		if mult > 1 {
			r.intervalMult = mult
		}
	}
}

// WithSleep sets the function used for sleeping.  By default, this is
// time.Sleep.
func WithSleep(sleep func(time.Duration)) Option {
	return func(r *retryConfig) {
		if sleep != nil {
			r.sleep = sleep
		}
	}
}

// WithMeasures provides a provider to use for metrics.
func WithMeasures(p provider.Provider) Option {
	return func(r *retryConfig) {
		if p != nil {
			r.measures = NewMeasures(p)
		}
	}
}

// RetryInsertService is a wrapper for a db.Inserter that attempts to insert
// a configurable number of times if the inserts fail.
type RetryInsertService struct {
	inserter db.Inserter
	config   retryConfig
}

// InsertRecords uses the inserter to insert the records and tries again if
// inserting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (ri RetryInsertService) InsertRecords(records ...db.Record) error {
	var err error

	retries := ri.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := ri.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ri.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.InsertType).Add(1.0)
			ri.config.sleep(sleepTime)
			sleepTime = sleepTime * ri.config.intervalMult
		}
		if err = ri.inserter.InsertRecords(records...); err == nil {
			break
		}
	}

	ri.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.InsertType).Add(1.0)
	return err
}

// CreateRetryInsertService takes an inserter and the options provided and
// creates a RetryInsertService.
func CreateRetryInsertService(inserter db.Inserter, options ...Option) RetryInsertService {
	ris := RetryInsertService{
		inserter: inserter,
		config: retryConfig{
			retries:      defaultRetries,
			interval:     defaultInterval,
			intervalMult: defaultIntervalMult,
			sleep:        defaultSleep,
		},
	}
	for _, o := range options {
		o(&ris.config)
	}
	return ris
}

// RetryUpdateService is a wrapper for a db.Pruner that attempts either part of
// the pruning process a configurable number of times.
type RetryUpdateService struct {
	pruner db.Pruner
	config retryConfig
}

// GetRecordsToDelete uses the pruner to get records and tries again if
// getting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (ru RetryUpdateService) GetRecordsToDelete(shard int, limit int, deathDate int64) ([]db.RecordToDelete, error) {
	var (
		err       error
		recordIDs []db.RecordToDelete
	)

	retries := ru.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := ru.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ru.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.ReadType).Add(1.0)
			ru.config.sleep(sleepTime)
			sleepTime = sleepTime * ru.config.intervalMult
		}
		if recordIDs, err = ru.pruner.GetRecordsToDelete(shard, limit, deathDate); err == nil {
			break
		}
	}

	ru.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return recordIDs, err
}

// DeleteRecord uses the pruner to delete a record and tries again if
// deleting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (ru RetryUpdateService) DeleteRecord(shard int, deathdate int64, recordID int64) error {
	var err error

	retries := ru.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := ru.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ru.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
			ru.config.sleep(sleepTime)
			sleepTime = sleepTime * ru.config.intervalMult
		}
		if err = ru.pruner.DeleteRecord(shard, deathdate, recordID); err == nil {
			break
		}
	}

	ru.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
	return err
}

// CreateRetryUpdateService takes a pruner and the options provided and creates
// a RetryUpdateService.
func CreateRetryUpdateService(pruner db.Pruner, options ...Option) RetryUpdateService {
	rus := RetryUpdateService{
		pruner: pruner,
		config: retryConfig{
			retries:      defaultRetries,
			interval:     defaultInterval,
			intervalMult: defaultIntervalMult,
			sleep:        defaultSleep,
		},
	}
	for _, o := range options {
		o(&rus.config)
	}
	return rus
}

// RetryListGService is a wrapper for a blacklist Updater that attempts to
// get the blacklist a configurable number of times if the gets fail.
type RetryListGService struct {
	lg     blacklist.Updater
	config retryConfig
}

// GetBlacklist uses the updater to get the blacklist and tries again if
// getting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (ltg RetryListGService) GetBlacklist() (list []blacklist.BlackListedItem, err error) {
	retries := ltg.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := ltg.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ltg.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.ListReadType).Add(1.0)
			ltg.config.sleep(sleepTime)
			sleepTime = sleepTime * ltg.config.intervalMult
		}
		if list, err = ltg.lg.GetBlacklist(); err == nil {
			break
		}
	}

	ltg.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.ListReadType).Add(1.0)
	return
}

// CreateRetryListGService takes an updater and the options provided and creates
// a RetryListGService.
func CreateRetryListGService(listGetter blacklist.Updater, options ...Option) RetryListGService {
	rlgs := RetryListGService{
		lg: listGetter,
		config: retryConfig{
			retries:      defaultRetries,
			interval:     defaultInterval,
			intervalMult: defaultIntervalMult,
			sleep:        defaultSleep,
		},
	}
	for _, o := range options {
		o(&rlgs.config)
	}
	return rlgs
}

// RetryRGService is a wrapper for a record getter that attempts to
// get records for a device a configurable number of times if the gets fail.
type RetryRGService struct {
	rg     db.RecordGetter
	config retryConfig
}

// GetRecords uses the getter to get records for a device and tries again if
// getting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (rtg RetryRGService) GetRecords(deviceID string, limit int) ([]db.Record, error) {
	var (
		err    error
		record []db.Record
	)

	retries := rtg.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := rtg.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.ReadType).Add(1.0)
			rtg.config.sleep(sleepTime)
			sleepTime = sleepTime * rtg.config.intervalMult
		}
		if record, err = rtg.rg.GetRecords(deviceID, limit); err == nil {
			break
		}
	}

	rtg.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return record, err
}

// GetRecordsOfType uses the getter to get records of a specified type for a
// device and tries again if getting fails.  Between each try, it calculates
// how long to wait and then waits for that period of time before trying again.
// Only the error from the last failure is returned.
func (rtg RetryRGService) GetRecordsOfType(deviceID string, limit int, eventType db.EventType) ([]db.Record, error) {
	var (
		err    error
		record []db.Record
	)

	retries := rtg.config.retries
	if retries < 1 {
		retries = 0
	}

	sleepTime := rtg.config.interval
	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.ReadType).Add(1.0)
			rtg.config.sleep(sleepTime)
			sleepTime = sleepTime * rtg.config.intervalMult
		}
		if record, err = rtg.rg.GetRecordsOfType(deviceID, limit, eventType); err == nil {
			break
		}
	}

	rtg.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return record, err
}

// CreateRetryRGService takes a record getter and the options provided and
// creates a RetryRGService.
func CreateRetryRGService(recordGetter db.RecordGetter, options ...Option) RetryRGService {
	rrgs := RetryRGService{
		rg: recordGetter,
		config: retryConfig{
			retries:      defaultRetries,
			interval:     defaultInterval,
			intervalMult: defaultIntervalMult,
			sleep:        defaultSleep,
		},
	}
	for _, o := range options {
		o(&rrgs.config)
	}
	return rrgs
}
