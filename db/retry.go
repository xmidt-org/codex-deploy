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
	InsertEvent(deviceID string, event Event, tombstoneKey string) error
}

type RetryInsertService struct {
	inserter Inserter
	config   retryConfig
}

func (ri RetryInsertService) InsertEvent(deviceID string, event Event, tombstoneKey string) error {
	return execute(ri.config, func() error {
		return ri.inserter.InsertEvent(deviceID, event, tombstoneKey)
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

type Updater interface {
	UpdateHistory(deviceID string, events []Event) error
}

type RetryUpdateService struct {
	updater Updater
	config  retryConfig
}

func (ru RetryUpdateService) UpdateHistory(deviceID string, events []Event) error {
	return execute(ru.config, func() error {
		return ru.updater.UpdateHistory(deviceID, events)
	})
}

func CreateRetryUpdateService(updater Updater, retries int, interval time.Duration) RetryUpdateService {
	return RetryUpdateService{
		updater: updater,
		config: retryConfig{
			retries:  retries,
			interval: interval,
			sleep:    time.Sleep,
		},
	}
}

type TombstoneGetter interface {
	GetTombstone(deviceID string) (map[string]Event, error)
}

type RetryTGService struct {
	tg       TombstoneGetter
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func (rtg RetryTGService) GetTombstone(deviceID string) (map[string]Event, error) {
	var (
		err       error
		tombstone map[string]Event
	)

	retries := rtg.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rtg.sleep(rtg.interval)
		}
		if tombstone, err = rtg.tg.GetTombstone(deviceID); err == nil {
			break
		}
	}

	return tombstone, err
}

func CreateRetryTGService(tombstoneGetter TombstoneGetter, retries int, interval time.Duration) RetryTGService {
	return RetryTGService{
		tg:       tombstoneGetter,
		retries:  retries,
		interval: interval,
		sleep:    time.Sleep,
	}
}

type HistoryGetter interface {
	GetHistory(deviceID string) (History, error)
}

type RetryHGService struct {
	hg       HistoryGetter
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func (rhg RetryHGService) GetHistory(deviceID string) (History, error) {
	var (
		err     error
		history History
	)

	retries := rhg.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			rhg.sleep(rhg.interval)
		}
		if history, err = rhg.hg.GetHistory(deviceID); err == nil {
			break
		}
	}

	return history, err
}

func CreateRetryHGService(historyGetter HistoryGetter, retries int, interval time.Duration) RetryHGService {
	return RetryHGService{
		hg:       historyGetter,
		retries:  retries,
		interval: interval,
		sleep:    time.Sleep,
	}
}
