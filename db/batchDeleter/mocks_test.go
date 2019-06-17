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

package batchDeleter

import (
	"github.com/Comcast/codex/db"
	"github.com/stretchr/testify/mock"
)

type mockPruner struct {
	mock.Mock
}

func (p *mockPruner) GetRecordsToDelete(shard int, limit int, deathDate int64) ([]db.RecordToDelete, error) {
	args := p.Called(shard, limit, deathDate)
	return args.Get(0).([]db.RecordToDelete), args.Error(1)
}

func (p *mockPruner) DeleteRecord(shard int, deathdate int64, recordID int64) error {
	args := p.Called(shard, deathdate, recordID)
	return args.Error(0)
}
