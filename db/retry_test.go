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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryInsertEvent(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description string
		numCalls    int
		retries     int
		finalError  error
		expectedErr error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Eventual Success",
			numCalls:    3,
			retries:     5,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description: "Eventual Failure",
			numCalls:    4,
			retries:     3,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockInserter)
			if tc.numCalls > 1 {
				mockObj.On("InsertEvent", mock.Anything, mock.Anything, mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("InsertEvent", mock.Anything, mock.Anything, mock.Anything).Return(tc.finalError).Once()
			}

			retryInsertService := RetryInsertService{
				inserter: mockObj,
				retries:  tc.retries,
				interval: interval,
				sleep: func(t time.Duration) {
					assert.Equal(interval, t)
				},
			}
			err := retryInsertService.InsertEvent("", Event{}, "")
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestCreateRetryInsertService(t *testing.T) {
	r := RetryInsertService{
		inserter: new(mockInserter),
		retries:  322,
		interval: 2 * time.Minute,
	}
	assert := assert.New(t)
	newService := CreateRetryInsertService(r.inserter, r.retries, r.interval)
	assert.Equal(r.inserter, newService.inserter)
	assert.Equal(r.retries, newService.retries)
	assert.Equal(r.interval, newService.interval)
}

func TestRetryUpdateHistory(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description string
		numCalls    int
		retries     int
		finalError  error
		expectedErr error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Eventual Success",
			numCalls:    3,
			retries:     5,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description: "Eventual Failure",
			numCalls:    4,
			retries:     3,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockUpdater)
			if tc.numCalls > 1 {
				mockObj.On("UpdateHistory", mock.Anything, mock.Anything, mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("UpdateHistory", mock.Anything, mock.Anything, mock.Anything).Return(tc.finalError).Once()
			}

			retryInsertService := RetryUpdateService{
				updater:  mockObj,
				retries:  tc.retries,
				interval: interval,
				sleep: func(t time.Duration) {
					assert.Equal(interval, t)
				},
			}
			err := retryInsertService.UpdateHistory("", []Event{})
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryUpdateService(t *testing.T) {
	r := RetryUpdateService{
		updater:  new(mockUpdater),
		retries:  322,
		interval: 2 * time.Minute,
	}
	assert := assert.New(t)
	newService := CreateRetryUpdateService(r.updater, r.retries, r.interval)
	assert.Equal(r.updater, newService.updater)
	assert.Equal(r.retries, newService.retries)
	assert.Equal(r.interval, newService.interval)
}

func TestRetryGetTombstone(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description string
		numCalls    int
		retries     int
		finalError  error
		expectedErr error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Eventual Success",
			numCalls:    3,
			retries:     5,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description: "Eventual Failure",
			numCalls:    4,
			retries:     3,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockTG)
			if tc.numCalls > 1 {
				mockObj.On("GetTombstone", mock.Anything, mock.Anything, mock.Anything).Return(map[string]Event{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetTombstone", mock.Anything, mock.Anything, mock.Anything).Return(map[string]Event{}, tc.finalError).Once()
			}

			retryTGService := RetryTGService{
				tg:       mockObj,
				retries:  tc.retries,
				interval: interval,
				sleep: func(t time.Duration) {
					assert.Equal(interval, t)
				},
			}
			_, err := retryTGService.GetTombstone("")
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryTGService(t *testing.T) {
	r := RetryTGService{
		tg:       new(mockTG),
		retries:  322,
		interval: 2 * time.Minute,
	}
	assert := assert.New(t)
	newService := CreateRetryTGService(r.tg, r.retries, r.interval)
	assert.Equal(r.tg, newService.tg)
	assert.Equal(r.retries, newService.retries)
	assert.Equal(r.interval, newService.interval)
}

func TestRetryGetHistory(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	interval := 8 * time.Second
	tests := []struct {
		description string
		numCalls    int
		retries     int
		finalError  error
		expectedErr error
	}{
		{
			description: "Initial Success",
			numCalls:    1,
			retries:     1,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Eventual Success",
			numCalls:    3,
			retries:     5,
			finalError:  nil,
			expectedErr: nil,
		},
		{
			description: "Initial Failure",
			numCalls:    1,
			retries:     0,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
		{
			description: "Eventual Failure",
			numCalls:    4,
			retries:     3,
			finalError:  failureErr,
			expectedErr: failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockHG)
			if tc.numCalls > 1 {
				mockObj.On("GetHistory", mock.Anything, mock.Anything, mock.Anything).Return(History{}, initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				mockObj.On("GetHistory", mock.Anything, mock.Anything, mock.Anything).Return(History{}, tc.finalError).Once()
			}

			retryInsertService := RetryHGService{
				hg:       mockObj,
				retries:  tc.retries,
				interval: interval,
				sleep: func(t time.Duration) {
					assert.Equal(interval, t)
				},
			}
			_, err := retryInsertService.GetHistory("")
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}

}

func TestCreateRetryHGService(t *testing.T) {
	r := RetryHGService{
		hg:       new(mockHG),
		retries:  322,
		interval: 2 * time.Minute,
	}
	assert := assert.New(t)
	newService := CreateRetryHGService(r.hg, r.retries, r.interval)
	assert.Equal(r.hg, newService.hg)
	assert.Equal(r.retries, newService.retries)
	assert.Equal(r.interval, newService.interval)
}
