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

type Inserter interface {
	InsertEvent(deviceID string, event Event, tombstoneKey string) error
}

type RetryInsertService struct {
	Inserter
	retries int
}

func (ri RetryInsertService) InsertEvent(deviceID string, event Event, tombstoneKey string) error {
	var err error

	retries := ri.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if err = ri.InsertEvent(deviceID, event, tombstoneKey); err == nil {
			break
		}
	}

	return err
}

func CreateRetryInsertService(i Inserter, r int) RetryInsertService {
	return RetryInsertService{i, r}
}

type Updater interface {
	UpdateHistory(deviceID string, events []Event) error
}

type RetryUpdateService struct {
	Updater
	retries int
}

func (ru RetryUpdateService) UpdateHistory(deviceID string, events []Event) error {
	var err error

	retries := ru.retries
	if retries < 1 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		if err = ru.UpdateHistory(deviceID, events); err == nil {
			break
		}
	}

	return err
}

func CreateRetryUpdateService(u Updater, r int) RetryUpdateService {
	return RetryUpdateService{u, r}
}
