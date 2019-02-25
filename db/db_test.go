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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	goodEvent = Event{
		Source:      "test source",
		Destination: "test destination",
		Details:     map[string]interface{}{"test key": "test value"},
	}
)

func TestGetRecords(t *testing.T) {
	tests := []struct {
		description     string
		deviceID        string
		expectedRecords []Record
		expectedErr     error
		expectedCalls   int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			expectedRecords: []Record{
				{
					ID:       1,
					Type:     0,
					DeviceID: "1234",
				},
			},
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			description:     "Invalid Device Error",
			deviceID:        "",
			expectedRecords: []Record{},
			expectedErr:     errInvaliddeviceID,
			expectedCalls:   0,
		},
		{
			description:     "Get Error",
			deviceID:        "1234",
			expectedRecords: []Record{},
			expectedErr:     errors.New("test Get error"),
			expectedCalls:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockFinder)
			dbConnection := Connection{
				finder: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshaledRecords, err := json.Marshal(tc.expectedRecords)
				assert.Nil(err)
				mockObj.On("find", mock.Anything, mock.Anything).Return(tc.expectedErr, marshaledRecords).Times(tc.expectedCalls)
			}
			records, err := dbConnection.GetRecords(tc.deviceID)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedRecords, records)
		})
	}
}

func TestGetRecordsOfType(t *testing.T) {
	tests := []struct {
		description     string
		deviceID        string
		eventType       int
		expectedRecords []Record
		expectedErr     error
		expectedCalls   int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			eventType:   1,
			expectedRecords: []Record{
				{
					ID:       1,
					Type:     1,
					DeviceID: "1234",
				},
			},
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			description:     "Invalid Event type Error",
			deviceID:        "",
			eventType:       -1,
			expectedRecords: []Record{},
			expectedErr:     errInvalidEventType,
			expectedCalls:   0,
		},
		{
			description:     "Invalid Device Error",
			deviceID:        "",
			expectedRecords: []Record{},
			expectedErr:     errInvaliddeviceID,
			expectedCalls:   0,
		},
		{
			description:     "Get Error",
			deviceID:        "1234",
			expectedRecords: []Record{},
			expectedErr:     errors.New("test Get error"),
			expectedCalls:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockFinder)
			dbConnection := Connection{
				finder: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshaledRecords, err := json.Marshal(tc.expectedRecords)
				assert.Nil(err)
				mockObj.On("find", mock.Anything, mock.Anything).Return(tc.expectedErr, marshaledRecords).Times(tc.expectedCalls)
			}
			records, err := dbConnection.GetRecordsOfType(tc.deviceID, tc.eventType)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedRecords, records)
		})
	}
}

func TestUpdateHistory(t *testing.T) {
	pruneTestErr := errors.New("test prune history error")
	tests := []struct {
		description string
		time        time.Time
		pruneErr    error
		expectedErr error
	}{
		{
			description: "Success",
			time:        time.Now(),
			pruneErr:    nil,
			expectedErr: nil,
		},
		{
			description: "Prune History Error",
			time:        time.Now(),
			pruneErr:    pruneTestErr,
			expectedErr: pruneTestErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockDeleter)
			dbConnection := Connection{
				deleter: mockObj,
			}
			mockObj.On("delete", mock.Anything, mock.Anything).Return(tc.pruneErr).Once()
			err := dbConnection.PruneRecords(tc.time)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestInsertEvent(t *testing.T) {
	testCreateErr := errors.New("test create error")
	goodRecord := Record{
		DeviceID: "1234",
	}

	tests := []struct {
		description   string
		record        Record
		createErr     error
		expectedErr   error
		expectedCalls int
	}{
		{
			description:   "Success",
			record:        goodRecord,
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			description:   "Invalid Event Error",
			record:        Record{},
			expectedErr:   errInvaliddeviceID,
			expectedCalls: 0,
		},
		{
			description:   "Create Error",
			record:        goodRecord,
			createErr:     testCreateErr,
			expectedErr:   testCreateErr,
			expectedCalls: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockCreator)
			dbConnection := Connection{
				creator: mockObj,
			}
			if tc.expectedCalls > 0 {
				mockObj.On("create", mock.Anything).Return(tc.createErr).Times(tc.expectedCalls)
			}
			err := dbConnection.InsertRecord(tc.record)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	tests := []struct {
		description string
		expectedErr error
	}{
		{
			description: "Success",
			expectedErr: nil,
		},
		{
			description: "Execute Error",
			expectedErr: errors.New("test delete error"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockDeleter)
			dbConnection := Connection{
				deleter: mockObj,
			}
			mockObj.On("delete", mock.Anything, mock.Anything).Return(tc.expectedErr).Once()
			err := dbConnection.RemoveAll()
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		description string
		expectedErr error
	}{
		{
			description: "Success",
			expectedErr: nil,
		},
		{
			description: "Close Error",
			expectedErr: errors.New("test close error"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockCloser)
			dbConnection := Connection{
				closer: mockObj,
			}
			mockObj.On("close").Return(tc.expectedErr).Once()
			err := dbConnection.Close()
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
