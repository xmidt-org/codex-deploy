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
	closer interface {
		close() error
	}
)

type dbDecorator struct {
	*gorm.DB
	measures    Measures
	stopThreads []chan struct{}
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
func (b *dbDecorator) close() error {
	for _, stopThread := range b.stopThreads {
		stopThread <- struct{}{}
	}

	return b.DB.Close()
}

func connect(connSpecStr string, maxIdleConns int, maxOpenCons int, measures Measures) (*dbDecorator, error) {
	c, err := gorm.Open("postgres", connSpecStr)

	if err != nil {
		return nil, err
	}

	db := &dbDecorator{c, measures, []chan struct{}{}}
	db.setupMetrics()
	db.configure(maxIdleConns, maxOpenCons)

	return db, nil
}

func (b *dbDecorator) configure(maxIdleConns int, maxOpenCons int, ) {
	if maxIdleConns < 2 {
		maxIdleConns = 2
	}
	b.DB.DB().SetMaxIdleConns(maxIdleConns)
	b.DB.DB().SetMaxOpenConns(maxOpenCons)
}

func (b *dbDecorator) setupMetrics() {
	// ping to check status
	pingStop := doEvery(time.Second, func() {
		err := b.DB.DB().Ping()
		if err != nil {
			b.measures.ConnectionStatus.Set(0.0)
		} else {
			b.measures.ConnectionStatus.Set(1.0)
		}
	})
	b.stopThreads = append(b.stopThreads, pingStop)

	// baseline
	startStats := b.DB.DB().Stats()
	prevWaitCount := startStats.WaitCount
	prevWaitDuration := startStats.WaitDuration.Nanoseconds()
	prevMaxIdleClosed := startStats.MaxIdleClosed
	prevMaxLifetimeClosed := startStats.MaxLifetimeClosed

	// update measurements
	metricsStop := doEvery(time.Second, func() {
		stats := b.DB.DB().Stats()

		// current connections
		b.measures.PoolOpenConnections.Set(float64(stats.OpenConnections))
		b.measures.PoolInUseConnections.Set(float64(stats.InUse))
		b.measures.PoolIdleConnections.Set(float64(stats.Idle))

		// Counters
		b.measures.SQLWaitCount.Add(float64(stats.WaitCount - prevWaitCount))
		b.measures.SQLWaitDuration.Add(float64(stats.WaitDuration.Nanoseconds() - prevWaitDuration))
		b.measures.SQLMaxIdleClosed.Add(float64(stats.MaxIdleClosed - prevMaxIdleClosed))
		b.measures.SQLMaxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed - prevMaxLifetimeClosed))
	})
	b.stopThreads = append(b.stopThreads, metricsStop)
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
