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

import "time"

type retryConfig struct {
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func execute(config retryConfig, op func() error) error {
	var err error

	retries := config.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			config.sleep(config.interval)
		}
		if err = op(); err == nil {
			break
		}
	}

	return err
}

type Inserter interface {
	InsertRecord(record Record) error
}

type RetryInsertService struct {
	inserter Inserter
	config   retryConfig
}

func (ri RetryInsertService) InsertRecord(record Record) error {
	return execute(ri.config, func() error {
		return ri.inserter.InsertRecord(record)
	})
}

func CreateRetryInsertService(inserter Inserter, retries int, interval time.Duration) RetryInsertService {
	return RetryInsertService{
		inserter: inserter,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
		},
	}
}

type Pruner interface {
	PruneRecords(t time.Time) error
}

type RetryUpdateService struct {
	pruner Pruner
	config retryConfig
}

func (ru RetryUpdateService) PruneRecords(t time.Time) error {
	return execute(ru.config, func() error {
		return ru.pruner.PruneRecords(t)
	})
}

func CreateRetryUpdateService(pruner Pruner, retries int, interval time.Duration) RetryUpdateService {
	return RetryUpdateService{
		pruner: pruner,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
		},
	}
}

type RecordGetter interface {
	GetRecords(deviceID string) ([]Record, error)
	GetRecordsOfType(deviceID string, eventType int) ([]Record, error)
}

type RetryRGService struct {
	rg       RecordGetter
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func (rtg RetryRGService) GetRecords(deviceID string) ([]Record, error) {
	var (
		err    error
		record []Record
	)

	retries := rtg.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.sleep(rtg.interval)
		}
		if record, err = rtg.rg.GetRecords(deviceID); err == nil {
			break
		}
	}

	return record, err
}

func (rtg RetryRGService) GetRecordsOfType(deviceID string, eventType int) ([]Record, error) {
	var (
		err    error
		record []Record
	)

	retries := rtg.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.sleep(rtg.interval)
		}
		if record, err = rtg.rg.GetRecordsOfType(deviceID, eventType); err == nil {
			break
		}
	}

	return record, err
}

func CreateRetryRGService(recordGetter RecordGetter, retries int, interval time.Duration) RetryRGService {
	return RetryRGService{
		rg:       recordGetter,
		retries:  retries,
		interval: interval,
		sleep:    time.Sleep,
	}
}
