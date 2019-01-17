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

// Interface describes the main functionality needed to connect to a db
type Interface interface {
	Initialize() error
	GetHistory(deviceID string) (History, error)
	GetTombstone(deviceID string) (map[string]Event, error)
	UpdateHistory(deviceID string, events []Event) error
	InsertEvent(deviceID string, event Event, tombstoneKey string) error
	RemoveAll() error
}

type bucketWrapper interface {
	Manager(username, password string) *gocb.BucketManager
	Get(key string, valuePtr interface{}) (gocb.Cas, error)
	MutateIn(key string, cas gocb.Cas, expiry uint32) *gocb.MutateInBuilder
	Counter(key string, delta, initial int64, expiry uint32) (uint64, gocb.Cas, error)
	Insert(key string, value interface{}, expiry uint32) (gocb.Cas, error)
	ExecuteN1qlQuery(q *gocb.N1qlQuery, params interface{}) (gocb.QueryResults, error)
}

// the prefixes for the different documents being stored in couchbase
const (
	historyDoc   = "history"
	counterDoc   = "counter"
	tombstoneDoc = "tombstone"
)

var (
	errInvaliddeviceID = errors.New("Invalid device ID")
	errInvalidEvent    = errors.New("Invalid event")
)

// TODO: Add a way to try to reconnect to the database after a command fails because the connection broke

// Connection contains the bucket connection and configuration values
type Connection struct {
	Server   string
	Username string
	Password string
	Bucket   string
	// number of times to try when initially connecting to the database
	NumRetries int
	// the time duration to add when creating TTLs for history documents
	Timeout    time.Duration
	bucketConn bucketWrapper
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
	ID string `json:"id"`

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
func (db *Connection) Initialize() error {
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
func (db *Connection) GetHistory(deviceID string) (History, error) {
	var (
		deviceInfo History
	)
	if deviceID == "" {
		return History{}, emperror.WrapWith(errors.New("Invalid device id"), "Get history not attempted",
			"device id", deviceID)
	}
	key := strings.Join([]string{historyDoc, deviceID}, ":")
	_, err := db.bucketConn.Get(key, &deviceInfo)
	if err != nil {
		return History{}, emperror.WrapWith(err, "Getting history from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// GetTombstone returns the tombstone (map of events) for a given device
func (db *Connection) GetTombstone(deviceID string) (map[string]Event, error) {
	var (
		deviceInfo map[string]Event
	)
	if deviceID == "" {
		return map[string]Event{}, emperror.WrapWith(errors.New("Invalid device id"), "Get tombstone not attempted",
			"device id", deviceID)
	}
	key := strings.Join([]string{tombstoneDoc, deviceID}, ":")
	_, err := db.bucketConn.Get(key, &deviceInfo)
	if err != nil {
		return map[string]Event{}, emperror.WrapWith(err, "Getting tombstone from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// UpdateHistory updates the history to the list of events given for a given device
func (db *Connection) UpdateHistory(deviceID string, events []Event) error {
	key := strings.Join([]string{historyDoc, deviceID}, ":")
	newTimeout := uint32(time.Now().Add(db.Timeout).Unix())
	_, err := db.bucketConn.MutateIn(key, 0, newTimeout).Upsert("events", &events, false).Execute()
	if err != nil {
		return emperror.WrapWith(err, "Update history failed", "device id", deviceID,
			"events", events)
	}
	return nil
}

// InsertEvent adds an event to the history of the given device id and adds it to the tombstone if a key is given
func (db *Connection) InsertEvent(deviceID string, event Event, tombstoneMapKey string) error {
	if valid, err := isEventValid(deviceID, event); !valid {
		return emperror.WrapWith(err, "Insert event not attempted", "device id", deviceID,
			"event", event)
	}
	// get event id using the device id
	counterKey := strings.Join([]string{counterDoc, deviceID}, ":")
	eventID, _, err := db.bucketConn.Counter(counterKey, 1, 0, 0)
	if err != nil {
		return emperror.WrapWith(err, "Failed to get event id", "device id", deviceID)
	}
	event.ID = strconv.FormatUint(eventID, 10)

	//if tombstonekey isn't empty string, then set the tombstone map at that key
	if tombstoneMapKey != "" {
		err = db.upsertToTombstone(deviceID, tombstoneMapKey, event)
		if err != nil {
			return err
		}
	}
	// append to the history, create if it doesn't exist
	err = db.upsertToHistory(deviceID, event)
	return err
}

func (db *Connection) upsertToTombstone(deviceID string, tombstoneMapKey string, event Event) error {
	tombstoneKey := strings.Join([]string{tombstoneDoc, deviceID}, ":")
	events := map[string]Event{tombstoneMapKey: event}
	_, err := db.bucketConn.Insert(tombstoneKey, &events, 0)
	if err != nil && err != gocb.ErrKeyExists {
		return emperror.WrapWith(err, "Failed to create tombstone", "device id", deviceID,
			"event id", event.ID, "event", event)
	}
	if err != nil {
		_, err = db.bucketConn.MutateIn(tombstoneKey, 0, 0).
			Upsert(tombstoneMapKey, &event, false).
			Execute()
		if err != nil {
			return emperror.WrapWith(err, "Failed to add event to tombstone", "device id", deviceID,
				"event id", event.ID, "event", event)
		}
	}
	return nil
}

func (db *Connection) upsertToHistory(deviceID string, event Event) error {
	newTimeout := uint32(time.Now().Add(db.Timeout).Unix())
	historyKey := strings.Join([]string{historyDoc, deviceID}, ":")
	eventDoc := History{
		Events: []Event{event},
	}
	_, err := db.bucketConn.Insert(historyKey, &eventDoc, newTimeout)
	if err != nil && err != gocb.ErrKeyExists {
		return emperror.WrapWith(err, "Failed to create history document", "device id", deviceID,
			"event id", event.ID, "event", event)
	}
	if err != nil {
		_, err = db.bucketConn.MutateIn(historyKey, 0, newTimeout).ArrayPrepend("events", &event, false).Execute()
		if err != nil {
			return emperror.WrapWith(err, "Failed to add event to history", "device id", deviceID,
				"event id", event.ID, "event", event)
		}
	}
	return nil
}

func isEventValid(deviceID string, event Event) (bool, error) {
	if deviceID == "" {
		return false, errInvaliddeviceID
	}
	if event.Source == "" || event.Destination == "" || len(event.Details) == 0 {
		return false, errInvalidEvent
	}
	return true, nil
}

// RemoveAll removes everything in the database.  Used for testing
func (db *Connection) RemoveAll() error {
	_, err := db.bucketConn.ExecuteN1qlQuery(gocb.NewN1qlQuery("DELETE FROM devices;"), nil)
	if err != nil {
		return emperror.Wrap(err, "Removing all devices from database failed")
	}
	return nil
}
