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
}

func (b *dbDecorator) find(out interface{}, where ...interface{}) error {
	db := b.Order("birth_date DESC").Find(out, where...)
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

func connect(connSpecStr string) (*dbDecorator, error) {
	c, err := gorm.Open("postgres", connSpecStr)
	return &dbDecorator{c}, err
}
