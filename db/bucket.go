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
	"gopkg.in/couchbase/gocb.v1"
)

type (
	historyPruner interface {
		pruneHistory(key string, expiry uint32, path string, value interface{}) error
	}

	historyModifier interface {
		create(key string, value interface{}, expiry uint32) error
		prependToHistory(key string, expiry uint32, path string, value interface{}) error
	}

	tombstoneModifier interface {
		create(key string, value interface{}, expiry uint32) error
		upsertTombstoneKey(key string, path string, value interface{}) error
	}

	idGenerator interface {
		getNextID(key string, delta, initial int64, expiry uint32) (uint64, error)
	}

	primaryIndexCreator interface {
		createPrimaryIndex(index string) error
	}

	docGetter interface {
		get(key string, valuePtr interface{}) error
	}

	n1qlExecuter interface {
		executeN1qlQuery(q *gocb.N1qlQuery, params interface{}) error
	}
)

type bucketDecorator struct {
	*gocb.Bucket
}

func (b *bucketDecorator) pruneHistory(key string, expiry uint32, path string, value interface{}) error {
	_, err := b.MutateIn(key, 0, expiry).Upsert(path, value, false).Execute()
	return err
}

func (b *bucketDecorator) create(key string, value interface{}, expiry uint32) error {
	_, err := b.Insert(key, value, expiry)
	return err
}

func (b *bucketDecorator) prependToHistory(key string, expiry uint32, path string, value interface{}) error {
	_, err := b.MutateIn(key, 0, expiry).
		ArrayPrepend(path, value, false).
		Execute()
	return err
}

func (b *bucketDecorator) upsertTombstoneKey(key string, path string, value interface{}) error {
	_, err := b.MutateIn(key, 0, 0).
		Upsert(path, value, false).
		Execute()
	return err
}

func (b *bucketDecorator) getNextID(key string, delta, initial int64, expiry uint32) (uint64, error) {
	val, _, err := b.Counter(key, delta, initial, expiry)
	return val, err
}

func (b *bucketDecorator) createPrimaryIndex(index string) error {
	return b.Manager("", "").CreatePrimaryIndex(index, true, false)
}

func (b *bucketDecorator) get(key string, valuePtr interface{}) error {
	_, err := b.Get(key, valuePtr)
	return err
}

func (b *bucketDecorator) executeN1qlQuery(q *gocb.N1qlQuery, params interface{}) error {
	_, err := b.ExecuteN1qlQuery(q, params)
	return err
}
