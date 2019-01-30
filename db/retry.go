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

type Inserter interface {
	InsertEvent(deviceID string, event Event, tombstoneKey string) error
}

type RetryInsertService struct {
	inserter Inserter
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func (ri RetryInsertService) InsertEvent(deviceID string, event Event, tombstoneKey string) error {
	var err error

	retries := ri.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ri.sleep(ri.interval)
		}
		if err = ri.inserter.InsertEvent(deviceID, event, tombstoneKey); err == nil {
			break
		}
	}

	return err
}

func CreateRetryInsertService(i Inserter, r int, d time.Duration) RetryInsertService {
	return RetryInsertService{
		inserter: i,
		retries:  r,
		interval: d,
		sleep:    time.Sleep,
	}
}

type Updater interface {
	UpdateHistory(deviceID string, events []Event) error
}

type RetryUpdateService struct {
	updater  Updater
	retries  int
	interval time.Duration
	sleep    func(time.Duration)
}

func (ru RetryUpdateService) UpdateHistory(deviceID string, events []Event) error {
	var err error

	retries := ru.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if i > 0 {
			ru.sleep(ru.interval)
		}
		if err = ru.updater.UpdateHistory(deviceID, events); err == nil {
			break
		}
	}

	return err
}

func CreateRetryUpdateService(u Updater, r int, d time.Duration) RetryUpdateService {
	return RetryUpdateService{
		updater:  u,
		retries:  r,
		interval: d,
		sleep:    time.Sleep,
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

func CreateRetryTGService(t TombstoneGetter, r int, d time.Duration) RetryTGService {
	return RetryTGService{
		tg:       t,
		retries:  r,
		interval: d,
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

func CreateRetryHGService(h HistoryGetter, r int, d time.Duration) RetryHGService {
	return RetryHGService{
		hg:       h,
		retries:  r,
		interval: d,
		sleep:    time.Sleep,
	}
}
