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

type Option func(r *retryConfig)

func WithRetries(retries int) Option {
	return func(r *retryConfig) {
		// only set retries if the value is valid
		if retries >= 0 {
			r.retries = retries
		}
	}
}

func WithInterval(interval time.Duration) Option {
	return func(r *retryConfig) {
		// only set interval if the value is valid
		if interval > time.Duration(0)*time.Second {
			r.interval = interval
		}
	}
}

func WithIntervalMultiplier(mult time.Duration) Option {
	return func(r *retryConfig) {
		if mult > 1 {
			r.intervalMult = mult
		}
	}
}

func WithSleep(sleep func(time.Duration)) Option {
	return func(r *retryConfig) {
		if sleep != nil {
			r.sleep = sleep
		}
	}
}

func WithMeasures(p provider.Provider) Option {
	return func(r *retryConfig) {
		if p != nil {
			r.measures = NewMeasures(p)
		}
	}
}

type RetryInsertService struct {
	inserter db.Inserter
	config   retryConfig
}

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

type RetryUpdateService struct {
	pruner db.Pruner
	config retryConfig
}

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

type RetryListGService struct {
	lg     blacklist.Updater
	config retryConfig
}

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

type RetryRGService struct {
	rg     db.RecordGetter
	config retryConfig
}

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
