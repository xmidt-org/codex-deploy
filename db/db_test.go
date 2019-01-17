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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/couchbase/gocb.v1"
	"testing"
)

var (
	goodEvent = Event{
		Source:      "test source",
		Destination: "test destination",
		Details:     map[string]interface{}{"test key": "test value"},
	}
)

func TestIsEventValid(t *testing.T) {
	tests := []struct {
		description      string
		deviceID         string
		event            Event
		expectedValidity bool
		expectedErr      error
	}{
		{
			description: "Success",
			deviceID:    "1234",
			event: Event{
				Source:      "test source",
				Destination: "test destination",
				Details:     map[string]interface{}{"test": "test value"},
			},
			expectedValidity: true,
			expectedErr:      nil,
		},
		{
			description:      "Empty device id error",
			deviceID:         "",
			event:            Event{},
			expectedValidity: false,
			expectedErr:      errInvaliddeviceID,
		},
		{
			description:      "Invalid event error",
			deviceID:         "1234",
			event:            Event{},
			expectedValidity: false,
			expectedErr:      errInvalidEvent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			valid, err := isEventValid(tc.deviceID, tc.event)
			assert.Equal(tc.expectedValidity, valid)
			assert.Equal(tc.expectedErr, err)
		})
	}

}

func TestGetHistory(t *testing.T) {
	var cas gocb.Cas
	tests := []struct {
		description     string
		deviceID        string
		expectedHistory History
		expectedErr     error
		expectedCalls   int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			expectedHistory: History{
				Events: []Event{
					{
						ID: "1234",
					},
				},
			},
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			description:     "Invalid Device Error",
			deviceID:        "",
			expectedHistory: History{},
			expectedErr:     errors.New("Invalid device id"),
			expectedCalls:   0,
		},
		{
			description:     "Get Error",
			deviceID:        "1234",
			expectedHistory: History{},
			expectedErr:     errors.New("test Get error"),
			expectedCalls:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockBucket)
			dbConnection := Connection{
				bucketConn: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshalledHistory, err := json.Marshal(tc.expectedHistory)
				assert.Nil(err)
				mockObj.On("Get", mock.Anything, mock.Anything).Return(cas, tc.expectedErr, marshalledHistory).Times(tc.expectedCalls)
			}
			history, err := dbConnection.GetHistory(tc.deviceID)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedHistory, history)
		})
	}
}

func TestGetTombstone(t *testing.T) {
	var cas gocb.Cas
	tests := []struct {
		description       string
		deviceID          string
		expectedTombstone map[string]Event
		expectedErr       error
		expectedCalls     int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			expectedTombstone: map[string]Event{
				"test": {
					ID: "1234",
				},
			},
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			description:       "Invalid Device Error",
			deviceID:          "",
			expectedTombstone: map[string]Event{},
			expectedErr:       errors.New("Invalid device id"),
			expectedCalls:     0,
		},
		{
			description:       "Get Error",
			deviceID:          "1234",
			expectedTombstone: map[string]Event{},
			expectedErr:       errors.New("test Get error"),
			expectedCalls:     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockBucket)
			dbConnection := Connection{
				bucketConn: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshalledTombstone, err := json.Marshal(tc.expectedTombstone)
				assert.Nil(err)
				mockObj.On("Get", mock.Anything, mock.Anything).Return(cas, tc.expectedErr, marshalledTombstone).Times(tc.expectedCalls)
			}
			tombstone, err := dbConnection.GetTombstone(tc.deviceID)
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedTombstone, tombstone)
		})
	}
}

func TestUpdateHistory(t *testing.T) {

}

func TestInsertEventSuccess(t *testing.T) {
	var cas gocb.Cas
	assert := assert.New(t)
	mockObj := new(mockBucket)
	dbConnection := Connection{
		bucketConn: mockObj,
	}
	mockObj.On("Counter", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(1), cas, nil).Once()
	mockObj.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(cas, nil).Twice()
	err := dbConnection.InsertEvent("1234", goodEvent, "test")
	mockObj.AssertExpectations(t)
	assert.Nil(err)
}

func TestInsertEventCounterFail(t *testing.T) {
	var cas gocb.Cas
	assert := assert.New(t)
	mockObj := new(mockBucket)
	dbConnection := Connection{
		bucketConn: mockObj,
	}
	expectedErr := errors.New("test counter fail")
	mockObj.On("Counter", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(1), cas, expectedErr).Once()
	err := dbConnection.InsertEvent("1234", goodEvent, "test")
	mockObj.AssertExpectations(t)
	assert.NotNil(err)
	if err != nil {
		assert.Contains(err.Error(), expectedErr.Error())
	}
}

func TestUpsertToTombstone(t *testing.T) {

}

func TestUpsertToHistory(t *testing.T) {

}
