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

package xvault

import (
	"github.com/stretchr/testify/mock"
)

type mockReader struct {
	mock.Mock
}

func (r *mockReader) read(key string) (map[string]interface{}, error) {
	args := r.Called(key)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

type mockAuthenticator struct {
	mock.Mock
}

func (a *mockAuthenticator) getAuth(path string, data map[string]interface{}) (*secretAuth, error) {
	args := a.Called(path, data)
	return args.Get(0).(*secretAuth), args.Error(1)
}
func (a *mockAuthenticator) setToken(token string) {
	a.Called(token)
	return
}
