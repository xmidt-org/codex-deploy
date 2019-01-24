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
	"errors"
	"testing"

	"github.com/spf13/viper"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestInitializeErrors(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		expectedErr error
	}{
		{
			description: "Empty Role/Secret Error",
			config:      Config{},
			expectedErr: ErrEmptyRoleSecretID,
		},
		{
			description: "Empty Base Path Error",
			config: Config{
				RoleID:   "test role id",
				SecretID: "test secret id",
			},
			expectedErr: ErrEmptyBasePath,
		},
		{
			description: "Success",
			config: Config{
				RoleID:     "test role id",
				SecretID:   "test secret id",
				BasePath:   "test base path",
				Address:    "test address",
				MaxRetries: 2,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			v := viper.New()
			v.Set("vault", &tc.config)
			assert.NotPanics(func() {
				c, err := Initialize(v)
				if tc.expectedErr != nil {
					assert.Nil(c)
				} else {
					assert.NotNil(c)
				}
				if tc.expectedErr == nil || err == nil {
					assert.Equal(tc.expectedErr, err)
				} else {
					assert.Contains(err.Error(), tc.expectedErr.Error())
				}
			})
		})
	}
}

func TestAuthenticate(t *testing.T) {
	testAuth := &secretAuth{&api.SecretAuth{}}
	testAuthErr := errors.New("test get auth error")
	tests := []struct {
		description    string
		getAuthResp    *secretAuth
		getAuthErr     error
		setTokenCalled bool
		expectedErr    error
	}{
		{
			description:    "Success",
			getAuthResp:    testAuth,
			getAuthErr:     nil,
			setTokenCalled: true,
			expectedErr:    nil,
		},
		{
			description: "Get Auth Error",
			getAuthResp: nil,
			getAuthErr:  testAuthErr,
			expectedErr: testAuthErr,
		},
		{
			description: "Nil Response Error",
			getAuthResp: nil,
			getAuthErr:  nil,
			expectedErr: errors.New("no auth info returned"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			mockObj := new(mockAuthenticator)
			mockObj.On("getAuth", mock.Anything, mock.Anything).Return(tc.getAuthResp, tc.getAuthErr).Once()
			if tc.setTokenCalled {
				mockObj.On("setToken", mock.Anything).Return().Once()
			}
			err := authenticate(mockObj, "", map[string]interface{}{})
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetUsernamePassword(t *testing.T) {
	tests := []struct {
		description      string
		stage            string
		key              string
		readCalled       bool
		readData         map[string]interface{}
		readErr          error
		expectedUsername string
		expectedPassword string
	}{
		{
			description:      "Values Success",
			stage:            "teststage",
			key:              "testkey",
			readCalled:       true,
			readData:         map[string]interface{}{"usr": "testusr", "pwd": "testpwd"},
			expectedUsername: "testusr",
			expectedPassword: "testpwd",
		},
		{
			description:      "Empty Success",
			key:              "testkey",
			readCalled:       true,
			readData:         map[string]interface{}{},
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			description:      "Empty Key",
			key:              "",
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			description:      "Read Error",
			key:              "testkey",
			readCalled:       true,
			readData:         map[string]interface{}{},
			readErr:          errors.New("test read error"),
			expectedUsername: "",
			expectedPassword: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockReader)
			client := Client{
				client: mockObj,
			}
			if tc.readCalled {
				mockObj.On("read", mock.Anything).Return(tc.readData, tc.readErr).Once()
			}
			username, pwd := client.GetUsernamePassword(tc.stage, tc.key)
			mockObj.AssertExpectations(t)
			assert.Equal(tc.expectedUsername, username)
			assert.Equal(tc.expectedPassword, pwd)
		})
	}
}
