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

/*
var (
	goodEvent = Event{
		Source:      "test source",
		Destination: "test destination",
		Details:     map[string]interface{}{"test key": "test value"},
	}
)

func TestOpenBucket(t *testing.T) {
	initialOpenBucketErr := errors.New("this error shouldn't be returned")
	authErr := errors.New("test authenticate error")
	openBucketErr := errors.New("test open bucket error")
	tests := []struct {
		description        string
		authenticateErr    error
		numOpenBucketCalls int
		openBucketErr      error
		expectedErr        error
	}{
		{
			description:        "Initial Success",
			authenticateErr:    nil,
			numOpenBucketCalls: 1,
			openBucketErr:      nil,
			expectedErr:        nil,
		},
		{
			description:        "Delayed Success",
			authenticateErr:    nil,
			numOpenBucketCalls: 4,
			openBucketErr:      nil,
			expectedErr:        nil,
		},
		{
			description:        "Authenticate Error",
			authenticateErr:    authErr,
			numOpenBucketCalls: 0,
			openBucketErr:      nil,
			expectedErr:        authErr,
		},
		{
			description:        "Open Bucket Error",
			authenticateErr:    nil,
			numOpenBucketCalls: 4,
			openBucketErr:      openBucketErr,
			expectedErr:        openBucketErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockCluster)
			dbConnection := Connection{
				numRetries:   3,
				waitTimeMult: 1,
			}
			mockObj.On("authenticate", mock.Anything).Return(tc.authenticateErr).Once()
			if tc.numOpenBucketCalls > 1 {
				mockObj.On("openBucket", mock.Anything).Return(initialOpenBucketErr).Times(tc.numOpenBucketCalls - 1)
			}
			if tc.numOpenBucketCalls > 0 {
				mockObj.On("openBucket", mock.Anything).Return(tc.openBucketErr).Once()
			}
			_, err := dbConnection.openBucket(mockObj, "", "", "")
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

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
			expectedErr:     errInvaliddeviceID,
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
			mockObj := new(mockDocGetter)
			dbConnection := Connection{
				docGetter: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshalledHistory, err := json.Marshal(tc.expectedHistory)
				assert.Nil(err)
				mockObj.On("get", mock.Anything, mock.Anything).Return(tc.expectedErr, marshalledHistory).Times(tc.expectedCalls)
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
			expectedErr:       errInvaliddeviceID,
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
			mockObj := new(mockDocGetter)
			dbConnection := Connection{
				docGetter: mockObj,
			}
			if tc.expectedCalls > 0 {
				marshalledTombstone, err := json.Marshal(tc.expectedTombstone)
				assert.Nil(err)
				mockObj.On("get", mock.Anything, mock.Anything).Return(tc.expectedErr, marshalledTombstone).Times(tc.expectedCalls)
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
	pruneTestErr := errors.New("test prune history error")
	tests := []struct {
		description string
		deviceID    string
		pruneErr    error
		expectedErr error
	}{
		{
			description: "Success",
			deviceID:    "1234",
			pruneErr:    nil,
			expectedErr: nil,
		},
		{
			description: "Invalid Device Error",
			deviceID:    "",
			pruneErr:    nil,
			expectedErr: errInvaliddeviceID,
		},
		{
			description: "Prune History Error",
			deviceID:    "1234",
			pruneErr:    pruneTestErr,
			expectedErr: pruneTestErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockHistoryPruner)
			dbConnection := Connection{
				historyPruner: mockObj,
			}
			if tc.deviceID != "" {
				mockObj.On("pruneHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.pruneErr).Once()
			}
			err := dbConnection.UpdateHistory(tc.deviceID, []Event{})
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
	testGetNextIDErr := errors.New("test get next id error")
	testCreateTombstoneErr := errors.New("test create tombstone error")
	testUpsertTombstoneErr := errors.New("test upsert tombstone error")
	testCreateHistoryErr := errors.New("test create history error")
	testPrependHistoryErr := errors.New("test upsert history error")
	goodEvent := Event{
		Source:      "test source",
		Destination: "test destination",
		Details:     map[string]interface{}{"test": ""},
	}

	tests := []struct {
		description     string
		deviceID        string
		event           Event
		tombstoneMapKey string

		getNextIDCalled       bool
		getNextIDErr          error
		createTombstoneCalled bool
		createTombstoneErr    error
		upsertTombstoneCalled bool
		upsertTombstoneErr    error
		createHistoryCalled   bool
		createHistoryErr      error
		prependHistoryCalled  bool
		prependHistoryErr     error

		expectedErr error
	}{
		{
			description:           "Success",
			deviceID:              "1234",
			event:                 goodEvent,
			tombstoneMapKey:       "test map key",
			getNextIDCalled:       true,
			createTombstoneCalled: true,
			createHistoryCalled:   true,
			expectedErr:           nil,
		},
		{
			description: "Invalid Event Error",
			deviceID:    "1234",
			event:       Event{},
			expectedErr: errInvalidEvent,
		},
		{
			description:     "Get Next ID Error",
			deviceID:        "1234",
			event:           goodEvent,
			getNextIDCalled: true,
			getNextIDErr:    testGetNextIDErr,
			expectedErr:     testGetNextIDErr,
		},
		{
			description:           "Create Tombstone Error",
			deviceID:              "1234",
			event:                 goodEvent,
			tombstoneMapKey:       "test map key",
			getNextIDCalled:       true,
			createTombstoneCalled: true,
			createTombstoneErr:    testCreateTombstoneErr,
			expectedErr:           testCreateTombstoneErr,
		},
		{
			description:           "Upsert Tombstone Error",
			deviceID:              "1234",
			event:                 goodEvent,
			tombstoneMapKey:       "test map key",
			getNextIDCalled:       true,
			createTombstoneCalled: true,
			createTombstoneErr:    gocb.ErrKeyExists,
			upsertTombstoneCalled: true,
			upsertTombstoneErr:    testUpsertTombstoneErr,
			expectedErr:           testUpsertTombstoneErr,
		},
		{
			description:         "Create History Error",
			deviceID:            "1234",
			event:               goodEvent,
			getNextIDCalled:     true,
			createHistoryCalled: true,
			createHistoryErr:    testCreateHistoryErr,
			expectedErr:         testCreateHistoryErr,
		},
		{
			description:          "Prepend History Error",
			deviceID:             "1234",
			event:                goodEvent,
			getNextIDCalled:      true,
			createHistoryCalled:  true,
			createHistoryErr:     gocb.ErrKeyExists,
			prependHistoryCalled: true,
			prependHistoryErr:    testPrependHistoryErr,
			expectedErr:          testPrependHistoryErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockIDGenerator := new(mockIDGenerator)
			mockTombstoneModifier := new(mockTombstoneModifier)
			mockHistoryModifier := new(mockHistoryModifier)
			dbConnection := Connection{
				timeout:           time.Second,
				tombstoneModifier: mockTombstoneModifier,
				idGenerator:       mockIDGenerator,
				historyModifier:   mockHistoryModifier,
			}

			// setup mock calls
			if tc.getNextIDCalled {
				mockIDGenerator.
					On("getNextID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(uint64(0), tc.getNextIDErr).
					Once()
			}
			if tc.createTombstoneCalled {
				mockTombstoneModifier.
					On("create", mock.Anything, mock.Anything, mock.Anything).
					Return(tc.createTombstoneErr).
					Once()
				if tc.upsertTombstoneCalled {
					mockTombstoneModifier.
						On("upsertTombstoneKey", mock.Anything, mock.Anything, mock.Anything).
						Return(tc.upsertTombstoneErr).
						Once()
				}
			}
			if tc.createHistoryCalled {
				mockHistoryModifier.
					On("create", mock.Anything, mock.Anything, mock.Anything).
					Return(tc.createHistoryErr).
					Once()
				if tc.prependHistoryCalled {
					mockHistoryModifier.
						On("prependToHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(tc.prependHistoryErr).
						Once()
				}
			}

			err := dbConnection.InsertEvent(tc.deviceID, tc.event, tc.tombstoneMapKey)
			if tc.getNextIDCalled {
				mockIDGenerator.AssertExpectations(t)
			}
			if tc.createTombstoneCalled {
				mockTombstoneModifier.AssertExpectations(t)
			}
			if tc.createHistoryCalled {
				mockHistoryModifier.AssertExpectations(t)
			}
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
			expectedErr: errors.New("test execute n1ql query error"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockN1QLExecuter)
			dbConnection := Connection{
				n1qlExecuter: mockObj,
			}
			mockObj.On("executeN1qlQuery", mock.Anything, mock.Anything).Return(tc.expectedErr).Once()
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
*/
