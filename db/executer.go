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
	// Import GORM-related packages.

	"database/sql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

type (
	finder interface {
		find(out interface{}, where ...interface{}) error
	}
	creator interface {
		create(value interface{}) error
	}
	deleter interface {
		delete(value interface{}, where ...interface{}) error
	}
	pinger interface {
		ping() error
	}
	closer interface {
		close() error
	}
	stats interface {
		getStats() sql.DBStats
	}
)

type dbDecorator struct {
	*gorm.DB
}

func (b *dbDecorator) find(out interface{}, where ...interface{}) error {
	db := b.Order("birth_date desc").Find(out, where...)
	return db.Error
}

func (b *dbDecorator) create(value interface{}) error {
	db := b.Create(value)
	return db.Error
}

func (b *dbDecorator) delete(value interface{}, where ...interface{}) error {
	db := b.Delete(value, where...)
	return db.Error
}

func (b *dbDecorator) ping() error {
	return b.DB.DB().Ping()
}

func (b *dbDecorator) close() error {
	return b.DB.Close()
}

func (b *dbDecorator) getStats() sql.DBStats {
	return b.DB.DB().Stats()
}

func connect(connSpecStr string) (*dbDecorator, error) {
	c, err := gorm.Open("postgres", connSpecStr)

	if err != nil {
		return nil, err
	}

	db := &dbDecorator{c}

	return db, nil
}

func doEvery(d time.Duration, f func()) chan struct{} {
	ticker := time.NewTicker(d)
	stop := make(chan struct{}, 1)
	go func(stop chan struct{}) {
		for {
			select {
			case <-ticker.C:
				f()
			case <-stop:
				return
			}
		}
	}(stop)
	return stop
}
