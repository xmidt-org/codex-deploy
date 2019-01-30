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
