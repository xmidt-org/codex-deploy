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
	"github.com/goph/emperror"
	"gopkg.in/couchbase/gocb.v1"
	"strconv"
	"strings"
	"time"
)

type Interface interface {
	Initialize() error
	GetHistory(deviceId string) (History, error)
	GetTombstone(deviceId string) (map[string]Event, error)
	UpdateHistory(deviceId string, events []Event) error
	InsertEvent(deviceId string, event Event, tombstoneKey string) error
	RemoveAll() error
}

// the prefixes for the different documents being stored in couchbase
const (
	historyDoc   = "history"
	counterDoc   = "counter"
	tombstoneDoc = "tombstone"
)

// TODO: Add a way to try to reconnect to the database after a command fails because the connection broke

// DbConnection contains the bucket connection and configuration values
type DbConnection struct {
	Server   string
	Username string
	Password string
	Bucket   string
	// number of times to try when initially connecting to the database
	NumRetries int
	// the time duration to add when creating TTLs for history documents
	Timeout    time.Duration
	bucketConn *gocb.Bucket
}

// History is a list of events related to a device id,
// and has a TTL
//
// swagger:model History
type History struct {
	// the list of events from newest to oldest
	Events []Event `json:"events"`
}

// Event represents the event information in the database
//
// swagger:model Event
type Event struct {
	// the id for the event
	//
	// required: true
	Id string `json:"id"`

	// the time this event was found
	//
	// required: true
	Time int64 `json:"time"`

	// the source of this event
	//
	// required: true
	Source string `json:"src"`

	// the destination of this event
	//
	// required: true
	Destination string `json:"dest"`

	// the partners related to this device
	//
	// required: true
	PartnerIDs []string `json:"partner_ids"`

	// the transaction id for this event
	//
	// required: true
	TransactionUUID string `json:"transaction_uuid,omitempty"`

	// payload
	//
	// required: false
	Payload []byte `json:"payload,omitempty"`

	// other metadata and details related to this state
	//
	// required: true
	Details map[string]interface{} `json:"details"`
}

// Initialize creates the connection with couchbase and opens the specified bucket
func (db *DbConnection) Initialize() error {
	var err error

	cluster, err := gocb.Connect("couchbase://" + db.Server)
	if err != nil {
		return emperror.WrapWith(err, "Connecting to couchbase failed", "server", db.Server)
	}
	// for verbose gocb logging when debugging
	//gocb.SetLogger(gocb.VerboseStdioLogger())
	err = cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: db.Username,
		Password: db.Password,
	})
	if err != nil {
		return emperror.WrapWith(err, "Couchbase authentication failed", "server", db.Server,
			"username", db.Username)
	}

	db.bucketConn, err = cluster.OpenBucket(db.Bucket, "")
	// retry if it fails
	waitTime := 1 * time.Second
	for attempt := 0; attempt < db.NumRetries && err != nil; attempt++ {
		time.Sleep(waitTime)
		db.bucketConn, err = cluster.OpenBucket(db.Bucket, "")
		waitTime = waitTime * 5
	}
	if err != nil {
		return emperror.WrapWith(err, "Opening bucket failed", "server", db.Server, "username", db.Username,
			"number of retries", db.NumRetries)
	}

	err = db.bucketConn.Manager("", "").CreatePrimaryIndex("", true, false)
	if err != nil {
		return emperror.Wrap(err, "Creating Primary Index failed")
	}

	return nil
}

// GetHistory returns the history (list of events) for a given device
func (db *DbConnection) GetHistory(deviceId string) (History, error) {
	var (
		deviceInfo History
	)
	if deviceId == "" {
		return History{}, emperror.WrapWith(errors.New("Invalid device id"), "Get history not attempted",
			"device id", deviceId)
	}
	key := strings.Join([]string{historyDoc, deviceId}, ":")
	_, err := db.bucketConn.Get(key, &deviceInfo)
	if err != nil {
		return History{}, emperror.WrapWith(err, "Getting history from database failed", "device id", deviceId)
	}
	return deviceInfo, nil
}

// GetTombstone returns the tombstone (map of events) for a given device
func (db *DbConnection) GetTombstone(deviceId string) (map[string]Event, error) {
	var (
		deviceInfo map[string]Event
	)
	if deviceId == "" {
		return map[string]Event{}, emperror.WrapWith(errors.New("Invalid device id"), "Get tombstone not attempted",
			"device id", deviceId)
	}
	key := strings.Join([]string{tombstoneDoc, deviceId}, ":")
	_, err := db.bucketConn.Get(key, &deviceInfo)
	if err != nil {
		return map[string]Event{}, emperror.WrapWith(err, "Getting tombstone from database failed", "device id", deviceId)
	}
	return deviceInfo, nil
}

// UpdateHistory updates the history to the list of events given for a given device
func (db *DbConnection) UpdateHistory(deviceId string, events []Event) error {
	key := strings.Join([]string{historyDoc, deviceId}, ":")
	newTimeout := uint32(time.Now().Add(db.Timeout).Unix())
	_, err := db.bucketConn.MutateIn(key, 0, newTimeout).Upsert("events", &events, false).Execute()
	if err != nil {
		return emperror.WrapWith(err, "Update history failed", "device id", deviceId,
			"events", events)
	}
	return nil
}

// InsertEvent adds an event to the history of the given device id and adds it to the tombstone if a key is given
func (db *DbConnection) InsertEvent(deviceId string, event Event, tombstoneMapKey string) error {
	if valid, err := isEventValid(deviceId, event); !valid {
		return emperror.WrapWith(err, "Insert event not attempted", "device id", deviceId,
			"event", event)
	}

	// get event id using the device id
	counterKey := strings.Join([]string{counterDoc, deviceId}, ":")
	eventID, _, err := db.bucketConn.Counter(counterKey, 1, 0, 0)
	if err != nil {
		return emperror.WrapWith(err, "Failed to get event id", "device id", deviceId)
	}

	event.Id = strconv.FormatUint(eventID, 10)

	//if tombstonekey isn't empty string, then set the tombstone map at that key
	if tombstoneMapKey != "" {
		tombstoneKey := strings.Join([]string{tombstoneDoc, deviceId}, ":")
		events := make(map[string]Event)
		events[tombstoneMapKey] = event
		_, err = db.bucketConn.Insert(tombstoneKey, &events, 0)
		if err != nil && err != gocb.ErrKeyExists {
			return emperror.WrapWith(err, "Failed to create tombstone", "device id", deviceId,
				"event id", eventID, "event", event)
		}
		if err != nil {
			_, err = db.bucketConn.MutateIn(tombstoneKey, 0, 0).
				Upsert(tombstoneMapKey, &event, false).
				Execute()
			if err != nil {
				return emperror.WrapWith(err, "Failed to add event to tombstone", "device id", deviceId,
					"event id", eventID, "event", event)
			}
		}
	}

	// append to the history, create if it doesn't exist
	newTimeout := uint32(time.Now().Add(db.Timeout).Unix())
	historyKey := strings.Join([]string{historyDoc, deviceId}, ":")
	eventDoc := History{
		Events: []Event{event},
	}
	_, err = db.bucketConn.Insert(historyKey, &eventDoc, newTimeout)
	if err != nil && err != gocb.ErrKeyExists {
		return emperror.WrapWith(err, "Failed to create history document", "device id", deviceId,
			"event id", eventID, "event", event)
	}
	if err != nil {
		_, err = db.bucketConn.MutateIn(historyKey, 0, newTimeout).ArrayPrepend("events", &event, false).Execute()
		if err != nil {
			return emperror.WrapWith(err, "Failed to add event to history", "device id", deviceId,
				"event id", eventID, "event", event)
		}
	}

	return nil
}

func isEventValid(deviceId string, event Event) (bool, error) {
	if deviceId == "" {
		return false, errors.New("Invalid device id")
	}
	if event.Source == "" || event.Destination == "" || len(event.Details) == 0 {
		return false, errors.New("Invalid event")
	}
	return true, nil
}

// RemoveAll removes everything in the database.  Used for testing
func (db *DbConnection) RemoveAll() error {
	_, err := db.bucketConn.ExecuteN1qlQuery(gocb.NewN1qlQuery("DELETE FROM devices;"), nil)
	if err != nil {
		return emperror.Wrap(err, "Removing all devices from database failed")
	}
	return nil
}
