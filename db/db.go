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
	"strconv"
	"strings"
	"time"

	"github.com/goph/emperror"
	"gopkg.in/couchbase/gocb.v1"
)

// Interface describes the main functionality needed to connect to a database.
type Interface interface {
	Initialize() error
	GetHistory(deviceID string) (History, error)
	GetTombstone(deviceID string) (map[string]Event, error)
	UpdateHistory(deviceID string, events []Event) error
	InsertEvent(deviceID string, event Event, tombstoneKey string) error
	RemoveAll() error
}

// These constants are prefixes for the different documents being stored in couchbase.
const (
	historyDoc   = "history"
	counterDoc   = "counter"
	tombstoneDoc = "tombstone"
)

var (
	errInvaliddeviceID = errors.New("Invalid device ID")
	errInvalidEvent    = errors.New("Invalid event")
)

// TODO: Possibly add a way to try to reconnect to the database after a command fails because the connection broke.

// Connection contains the tools to edit the database.
type Connection struct {
	// Number of times to try when  connecting to the database
	numRetries int
	// Multiplier of the wait time so that we can wait longer after each failure
	waitTimeMult time.Duration
	// The time duration to add when creating TTLs for history documents
	timeout           time.Duration
	historyPruner     historyPruner
	historyModifier   historyModifier
	tombstoneModifier tombstoneModifier
	idGenerator       idGenerator
	docGetter         docGetter
	n1qlExecuter      n1qlExecuter
}

// History is a list of events related to a device id.  It has a TTL.
//
// swagger:model History
type History struct {
	// the list of events from newest to oldest
	Events []Event `json:"events"`
}

// Tombstone is a map of events related to a device id.  It has no TTL.
//
// swagger:model Tombstone
type Tombstone map[string]Event

// Event represents the event information in the database.  It has no TTL.
//
// swagger:model Event
type Event struct {
	// The id for the event.
	//
	// required: true
	ID string `json:"id"`

	// The time this event was found.
	//
	// required: true
	Time int64 `json:"time"`

	// The source of this event.
	//
	// required: true
	Source string `json:"src"`

	// The destination of this event.
	//
	// required: true
	Destination string `json:"dest"`

	// The partners related to this device.
	//
	// required: true
	PartnerIDs []string `json:"partner_ids"`

	// The transaction id for this event.
	//
	// required: true
	TransactionUUID string `json:"transaction_uuid,omitempty"`

	// payload
	//
	// required: false
	Payload []byte `json:"payload,omitempty"`

	// Other metadata and details related to this state.
	//
	// required: true
	Details map[string]interface{} `json:"details"`
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(server, username, password, bucket string, numRetries int, timeout time.Duration) (*Connection, error) {
	db := Connection{
		timeout:      timeout,
		numRetries:   numRetries,
		waitTimeMult: 5,
	}
	cluster, err := connect("couchbase://" + server)
	if err != nil {
		return &Connection{}, emperror.WrapWith(err, "Connecting to couchbase failed", "server", server)
	}

	// for verbose gocb logging when debugging
	//gocb.SetLogger(gocb.VerboseStdioLogger())

	bucketConn, err := db.openBucket(cluster, username, password, bucket)
	if err != nil {
		return &Connection{}, emperror.With(err, "server", server)
	}

	db.historyPruner = bucketConn
	db.historyModifier = bucketConn
	db.tombstoneModifier = bucketConn
	db.idGenerator = bucketConn
	db.n1qlExecuter = bucketConn
	db.docGetter = bucketConn

	err = bucketConn.createPrimaryIndex("")
	if err != nil {
		return nil, emperror.Wrap(err, "Creating Primary Index failed")
	}
	return &db, nil
}

// OpenBucket creates the connection with couchbase and opens the specified bucket.
func (db *Connection) openBucket(cluster cluster, username, password, bucket string) (*bucketDecorator, error) {
	var err error
	err = cluster.authenticate(gocb.PasswordAuthenticator{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, emperror.WrapWith(err, "Couchbase authentication failed", "username", username)
	}

	bucketConn, err := cluster.openBucket(bucket)
	// retry if it fails
	waitTime := 1 * time.Second
	for attempt := 0; attempt < db.numRetries && err != nil; attempt++ {
		time.Sleep(waitTime)
		bucketConn, err = cluster.openBucket(bucket)
		waitTime = waitTime * db.waitTimeMult
	}
	if err != nil {
		return nil, emperror.WrapWith(err, "Opening bucket failed", "username", username,
			"number of retries", db.numRetries)
	}

	return bucketConn, nil
}

// GetHistory returns the history (list of events) for a given device.
func (db *Connection) GetHistory(deviceID string) (History, error) {
	var (
		deviceInfo History
	)
	if deviceID == "" {
		return History{}, emperror.WrapWith(errInvaliddeviceID, "Get history not attempted",
			"device id", deviceID)
	}
	key := strings.Join([]string{historyDoc, deviceID}, ":")
	err := db.docGetter.get(key, &deviceInfo)
	if err != nil {
		return History{}, emperror.WrapWith(err, "Getting history from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// GetTombstone returns the tombstone (map of events) for a given device.
func (db *Connection) GetTombstone(deviceID string) (map[string]Event, error) {
	var (
		deviceInfo map[string]Event
	)
	if deviceID == "" {
		return map[string]Event{}, emperror.WrapWith(errInvaliddeviceID, "Get tombstone not attempted",
			"device id", deviceID)
	}
	key := strings.Join([]string{tombstoneDoc, deviceID}, ":")
	err := db.docGetter.get(key, &deviceInfo)
	if err != nil {
		return map[string]Event{}, emperror.WrapWith(err, "Getting tombstone from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// UpdateHistory updates the history to the list of events given for a given device.
func (db *Connection) UpdateHistory(deviceID string, events []Event) error {
	if deviceID == "" {
		return emperror.WrapWith(errInvaliddeviceID, "Update history not attempted",
			"device id", deviceID)
	}
	key := strings.Join([]string{historyDoc, deviceID}, ":")
	newTimeout := uint32(time.Now().Add(db.timeout).Unix())
	err := db.historyPruner.pruneHistory(key, newTimeout, "events", &events)
	if err != nil {
		return emperror.WrapWith(err, "Update history failed", "device id", deviceID,
			"events", events)
	}
	return nil
}

// InsertEvent adds an event to the history of the given device id and adds it to the tombstone if a key is given.
func (db *Connection) InsertEvent(deviceID string, event Event, tombstoneMapKey string) error {
	if valid, err := isEventValid(deviceID, event); !valid {
		return emperror.WrapWith(err, "Insert event not attempted", "device id", deviceID,
			"event", event)
	}
	// get event id using the device id
	counterKey := strings.Join([]string{counterDoc, deviceID}, ":")
	eventID, err := db.idGenerator.getNextID(counterKey, 1, 0, 0)
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
	err := db.tombstoneModifier.create(tombstoneKey, &events, 0)
	if err != nil && err != gocb.ErrKeyExists {
		return emperror.WrapWith(err, "Failed to create tombstone", "device id", deviceID,
			"event id", event.ID, "event", event)
	}
	if err != nil {
		err = db.tombstoneModifier.upsertTombstoneKey(tombstoneKey, tombstoneMapKey, &event)
		if err != nil {
			return emperror.WrapWith(err, "Failed to add event to tombstone", "device id", deviceID,
				"event id", event.ID, "event", event)
		}
	}
	return nil
}

func (db *Connection) upsertToHistory(deviceID string, event Event) error {
	newTimeout := uint32(time.Now().Add(db.timeout).Unix())
	historyKey := strings.Join([]string{historyDoc, deviceID}, ":")
	eventDoc := History{
		Events: []Event{event},
	}
	err := db.historyModifier.create(historyKey, &eventDoc, newTimeout)
	if err != nil && err != gocb.ErrKeyExists {
		return emperror.WrapWith(err, "Failed to create history document", "device id", deviceID,
			"event id", event.ID, "event", event)
	}
	if err != nil {
		err = db.historyModifier.prependToHistory(historyKey, newTimeout, "events", &event)
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

// RemoveAll removes everything in the database.  Used for testing.
func (db *Connection) RemoveAll() error {
	err := db.n1qlExecuter.executeN1qlQuery(gocb.NewN1qlQuery("DELETE FROM devices"), nil)
	if err != nil {
		return emperror.Wrap(err, "Removing all devices from database failed")
	}
	return nil
}
