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
)

type dbDecorator struct {
	*gorm.DB
	measures Measures
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

func connect(connSpecStr string, maxIdleConns int, maxOpenCons int, measures Measures) (*dbDecorator, error) {
	c, err := gorm.Open("postgres", connSpecStr)

	if err == nil && c != nil {
		if maxIdleConns < 2 {
			maxIdleConns = 2
		}
		c.DB().SetMaxIdleConns(maxIdleConns)
		c.DB().SetMaxOpenConns(maxOpenCons)

		// ping to check status
		doEvery(time.Second, func() {
			err := c.DB().Ping()
			if err != nil {
				measures.ConnectionStatus.Set(0.0)
			} else {
				measures.ConnectionStatus.Set(1.0)
			}
		})

		// baseline
		startStats := c.DB().Stats()
		prevWaitCount := startStats.WaitCount
		prevWaitDuration := startStats.WaitDuration.Nanoseconds()
		prevMaxIdleClosed := startStats.MaxIdleClosed
		prevMaxLifetimeClosed := startStats.MaxLifetimeClosed

		// update measurements
		doEvery(time.Second, func() {
			stats := c.DB().Stats()

			// current connections
			measures.PoolOpenConnections.Set(float64(stats.OpenConnections))
			measures.PoolInUseConnections.Set(float64(stats.InUse))
			measures.PoolIdleConnections.Set(float64(stats.Idle))

			// Counters
			measures.SQLWaitCount.Add(float64(stats.WaitCount - prevWaitCount))
			measures.SQLWaitDuration.Add(float64(stats.WaitDuration.Nanoseconds() - prevWaitDuration))
			measures.SQLMaxIdleClosed.Add(float64(stats.MaxIdleClosed - prevMaxIdleClosed))
			measures.SQLMaxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed - prevMaxLifetimeClosed))
		})
	}

	return &dbDecorator{c, measures}, err
}

func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}
}
