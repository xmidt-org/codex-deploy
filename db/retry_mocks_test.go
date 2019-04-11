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
	"github.com/stretchr/testify/mock"
)

type mockInserter struct {
	mock.Mock
}

func (i *mockInserter) InsertRecords(records ...Record) error {
	args := i.Called(records)
	return args.Error(0)
}

type mockPruner struct {
	mock.Mock
}

func (p *mockPruner) PruneRecords(t int64) error {
	args := p.Called(t)
	return args.Error(0)
}

type mockRG struct {
	mock.Mock
}

func (rg *mockRG) GetRecords(deviceID string, limit int) ([]Record, error) {
	args := rg.Called(deviceID, limit)
	return args.Get(0).([]Record), args.Error(1)
}

func (rg *mockRG) GetRecordsOfType(deviceID string, limit int, eventType int) ([]Record, error) {
	args := rg.Called(deviceID, limit, eventType)
	return args.Get(0).([]Record), args.Error(1)
}

type mockLG struct {
	mock.Mock
}

func (rg *mockLG) GetBlacklist() ([]BlackDevice, error) {
	args := rg.Called()
	return args.Get(0).([]BlackDevice), args.Error(1)
}
