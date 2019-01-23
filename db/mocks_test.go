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

type mockCluster struct {
	mock.Mock
}

func (c *mockCluster) authenticate(auth gocb.Authenticator) error {
	args := c.Called(auth)
	return args.Error(0)
}

func (c *mockCluster) openBucket(bucket string) (*bucketDecorator, error) {
	args := c.Called(bucket)
	return &bucketDecorator{}, args.Error(0)
}

type mockHistoryPruner struct {
	mock.Mock
}

func (hp *mockHistoryPruner) pruneHistory(key string, expiry uint32, path string, value interface{}) error {
	args := hp.Called(key, expiry, path, value)
	return args.Error(0)
}

type mockHistoryModifier struct {
	mock.Mock
}

func (hm *mockHistoryModifier) create(key string, value interface{}, expiry uint32) error {
	args := hm.Called(key, value, expiry)
	return args.Error(0)
}
func (hm *mockHistoryModifier) prependToHistory(key string, expiry uint32, path string, value interface{}) error {
	args := hm.Called(key, expiry, path, value)
	return args.Error(0)
}

type mockTombstoneModifier struct {
	mock.Mock
}

func (tm *mockTombstoneModifier) create(key string, value interface{}, expiry uint32) error {
	args := tm.Called(key, value, expiry)
	return args.Error(0)
}
func (tm *mockTombstoneModifier) upsertTombstoneKey(key string, path string, value interface{}) error {
	args := tm.Called(key, path, value)
	return args.Error(0)
}

type mockIDGenerator struct {
	mock.Mock
}

func (ig *mockIDGenerator) getNextID(key string, delta, initial int64, expiry uint32) (uint64, error) {
	args := ig.Called(key, delta, initial, expiry)
	return args.Get(0).(uint64), args.Error(1)
}

type mockPrimaryIndexCreator struct {
	mock.Mock
}

func (pic *mockPrimaryIndexCreator) createPrimaryIndex(index string) error {
	args := pic.Called(index)
	return args.Error(0)
}

type mockDocGetter struct {
	mock.Mock
}

func (dg *mockDocGetter) get(key string, valuePtr interface{}) error {
	args := dg.Called(key, valuePtr)
	err := json.Unmarshal(args.Get(1).([]byte), valuePtr)
	if err != nil {
		return err
	}
	return args.Error(0)
}

type mockN1QLExecuter struct {
	mock.Mock
}

func (ne *mockN1QLExecuter) executeN1qlQuery(q *gocb.N1qlQuery, params interface{}) error {
	args := ne.Called(q, params)
	return args.Error(0)
}

/*type mockBucket struct {
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
}*/
