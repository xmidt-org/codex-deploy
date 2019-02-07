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
	"database/sql"

	_ "github.com/lib/pq"
)

type (
	executer interface {
		execute(query string, args ...interface{}) error
	}
	enquirer interface {
		query(obj interface{}, query string, args ...interface{}) (*sql.Rows, error)
	}
)

type dbDecorator struct {
	*sql.DB
}

func (b *dbDecorator) execute(query string, args ...interface{}) error {
	_, err := b.Exec(query, args...)
	return err
}

func (b *dbDecorator) query(obj interface{}, query string, args ...interface{}) (*sql.Rows, error) {
	return b.Query(query, args...)
}

func connect(connSpecStr string) (*dbDecorator, error) {
	c, err := sql.Open("postgres", connSpecStr)
	return &dbDecorator{c}, err
}
