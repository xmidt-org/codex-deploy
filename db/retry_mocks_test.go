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

/*
import (
	"github.com/stretchr/testify/mock"
)

type mockInserter struct {
	mock.Mock
}

func (i *mockInserter) InsertEvent(deviceID string, event Event, tombstoneKey string) error {
	args := i.Called(deviceID, event, tombstoneKey)
	return args.Error(0)
}

type mockUpdater struct {
	mock.Mock
}

func (u *mockUpdater) UpdateHistory(deviceID string, events []Event) error {
	args := u.Called(deviceID, events)
	return args.Error(0)
}

type mockTG struct {
	mock.Mock
}

func (u *mockTG) GetTombstone(deviceID string) (map[string]Event, error) {
	args := u.Called(deviceID)
	return args.Get(0).(map[string]Event), args.Error(1)
}

type mockHG struct {
	mock.Mock
}

func (u *mockHG) GetHistory(deviceID string) (History, error) {
	args := u.Called(deviceID)
	return args.Get(0).(History), args.Error(1)
}
*/
