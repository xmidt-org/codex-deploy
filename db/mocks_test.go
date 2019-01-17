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
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"gopkg.in/couchbase/gocb.v1"
)

type mockBucket struct {
	mock.Mock
}

func (m *mockBucket) Manager(username, password string) *gocb.BucketManager {
	args := m.Called(username, password)
	return args.Get(0).(*gocb.BucketManager)
}

func (m *mockBucket) Get(key string, valuePtr interface{}) (gocb.Cas, error) {
	args := m.Called(key, valuePtr)
	json.Unmarshal(args.Get(2).([]byte), valuePtr)
	return args.Get(0).(gocb.Cas), args.Error(1)
}

func (m *mockBucket) MutateIn(key string, cas gocb.Cas, expiry uint32) *gocb.MutateInBuilder {
	args := m.Called(key, cas, expiry)
	return args.Get(0).(*gocb.MutateInBuilder)
}

func (m *mockBucket) Counter(key string, delta, initial int64, expiry uint32) (uint64, gocb.Cas, error) {
	args := m.Called(key, delta, initial, expiry)
	return args.Get(0).(uint64), args.Get(1).(gocb.Cas), args.Error(2)
}

func (m *mockBucket) Insert(key string, value interface{}, expiry uint32) (gocb.Cas, error) {
	args := m.Called(key, value, expiry)
	return args.Get(0).(gocb.Cas), args.Error(1)
}

func (m *mockBucket) ExecuteN1qlQuery(q *gocb.N1qlQuery, params interface{}) (gocb.QueryResults, error) {
	args := m.Called(q, params)
	return args.Get(0).(gocb.QueryResults), args.Error(1)
}
