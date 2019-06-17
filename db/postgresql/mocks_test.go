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

package postgresql

import (
	"encoding/json"

	"github.com/Comcast/codex/db"
	"github.com/stretchr/testify/mock"
)

type mockFinder struct {
	mock.Mock
}

func (f *mockFinder) findRecords(out *[]db.Record, limit int, where ...interface{}) error {
	args := f.Called(out, limit, where)
	err := json.Unmarshal(args.Get(1).([]byte), out)
	if err != nil {
		return err
	}
	return args.Error(0)
}

func (f *mockFinder) findRecordsToDelete(limit int, shard int, deathDate int64) ([]db.RecordToDelete, error) {
	args := f.Called(limit, shard, deathDate)
	return args.Get(0).([]db.RecordToDelete), args.Error(1)
}

type mockMultiInsert struct {
	mock.Mock
}

func (c *mockMultiInsert) insert(records []db.Record) (int64, error) {
	args := c.Called(records)
	return int64(args.Int(0)), args.Error(1)
}

type mockDeleter struct {
	mock.Mock
}

func (d *mockDeleter) delete(value *db.Record, limit int, where ...interface{}) (int64, error) {
	args := d.Called(value, limit, where)
	return int64(args.Int(0)), args.Error(1)
}

type mockCloser struct {
	mock.Mock
}

func (d *mockCloser) close() error {
	args := d.Called()
	return args.Error(0)
}

type mockPing struct {
	mock.Mock
}

func (d *mockPing) ping() error {
	args := d.Called()
	return args.Error(0)
}
