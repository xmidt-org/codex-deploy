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
)

type mockFinder struct {
	mock.Mock
}

func (f *mockFinder) find(out interface{}, where ...interface{}) error {
	args := f.Called(out, where)
	err := json.Unmarshal(args.Get(1).([]byte), out)
	if err != nil {
		return err
	}
	return args.Error(0)
}

type mockCreator struct {
	mock.Mock
}

func (c *mockCreator) create(value interface{}) error {
	args := c.Called(value)
	return args.Error(0)
}

type mockDeleter struct {
	mock.Mock
}

func (d *mockDeleter) delete(value interface{}, where ...interface{}) error {
	args := d.Called(value, where)
	return args.Error(0)
}
