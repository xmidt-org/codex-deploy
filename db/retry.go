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
	"github.com/Comcast/codex/blacklist"
	"time"

	"github.com/go-kit/kit/metrics/provider"
)

type retryConfig struct {
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
	measures Measures
}

type Inserter interface {
	InsertRecords(records ...Record) error
}

type RetryInsertService struct {
	inserter Inserter
	config   retryConfig
}

func (ri RetryInsertService) InsertRecords(records ...Record) error {
	var err error

	retries := ri.config.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ri.config.measures.SQLQueryRetryCount.With(typeLabel, insertType).Add(1.0)
			ri.config.sleep(ri.config.interval)
		}
		if err = ri.inserter.InsertRecords(records...); err == nil {
			break
		}
	}

	return err
}

func CreateRetryInsertService(inserter Inserter, retries int, interval time.Duration, provider provider.Provider) RetryInsertService {
	return RetryInsertService{
		inserter: inserter,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
			measures: NewMeasures(provider),
		},
	}
}

type Pruner interface {
	PruneRecords(t int64) error
}

type RetryUpdateService struct {
	pruner Pruner
	config retryConfig
}

func (ru RetryUpdateService) PruneRecords(t int64) error {
	var err error

	retries := ru.config.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ru.config.measures.SQLQueryRetryCount.With(typeLabel, deleteType).Add(1.0)
			ru.config.sleep(ru.config.interval)
		}
		if err = ru.pruner.PruneRecords(t); err == nil {
			break
		}
	}

	return err
}

func CreateRetryUpdateService(pruner Pruner, retries int, interval time.Duration, provider provider.Provider) RetryUpdateService {
	return RetryUpdateService{
		pruner: pruner,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
			measures: NewMeasures(provider),
		},
	}
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

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ltg.config.measures.SQLQueryRetryCount.With(typeLabel, listReadType).Add(1.0)
			ltg.config.sleep(ltg.config.interval)
		}
		if list, err = ltg.lg.GetBlacklist(); err == nil {
			break
		}
	}

	return
}

func CreateRetryListGService(listGetter blacklist.Updater, retries int, interval time.Duration, provider provider.Provider) RetryListGService {
	return RetryListGService{
		lg: listGetter,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
			measures: NewMeasures(provider),
		},
	}
}

type RecordGetter interface {
	GetRecords(deviceID string, limit int) ([]Record, error)
	GetRecordsOfType(deviceID string, limit int, eventType int) ([]Record, error)
}

type RetryRGService struct {
	rg     RecordGetter
	config retryConfig
}

func (rtg RetryRGService) GetRecords(deviceID string, limit int) ([]Record, error) {
	var (
		err    error
		record []Record
	)

	retries := rtg.config.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.config.measures.SQLQueryRetryCount.With(typeLabel, readType).Add(1.0)
			rtg.config.sleep(rtg.config.interval)
		}
		if record, err = rtg.rg.GetRecords(deviceID, limit); err == nil {
			break
		}
	}

	return record, err
}

func (rtg RetryRGService) GetRecordsOfType(deviceID string, limit int, eventType int) ([]Record, error) {
	var (
		err    error
		record []Record
	)

	retries := rtg.config.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.config.measures.SQLQueryRetryCount.With(typeLabel, readType).Add(1.0)
			rtg.config.sleep(rtg.config.interval)
		}
		if record, err = rtg.rg.GetRecordsOfType(deviceID, limit, eventType); err == nil {
			break
		}
	}

	return record, err
}

func CreateRetryRGService(recordGetter RecordGetter, retries int, interval time.Duration, provider provider.Provider) RetryRGService {
	return RetryRGService{
		rg: recordGetter,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
			measures: NewMeasures(provider),
		},
	}
}
