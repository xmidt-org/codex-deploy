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
	GetHistory(deviceId string) ([]Event, error)
	GetTombstone(deviceId string) (map[string]Event, error)
	RemoveHistory(deviceId string, numToRemove int) error
	InsertEvent(deviceId string, event Event, tombstoneKey string) error
	RemoveAll() error
}

const (
	historyDoc   = "history"
	counterDoc   = "counter"
	tombstoneDoc = "tombstone"
)

type DbConnection struct {
	Server     string
	Username   string
	Password   string
	Bucket     string
	NumRetries int
	Timeout    time.Duration
	bucketConn *gocb.Bucket
}

// Tombstone hold the map of the last of certain events
// that are saved so that they are not deleted
//
// swagger:model Tombstone
type Tombstone struct {
	Events map[string]Event `json:"events"`
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

func (db *DbConnection) GetHistory(deviceId string) ([]Event, error) {
	var (
		deviceInfo []Event
	)
	if deviceId == "" {
		return []Event{}, emperror.WrapWith(errors.New("Invalid device id"), "Get history not attempted",
			"device id", deviceId)
	}
	key := strings.Join([]string{historyDoc, deviceId}, ":")
	_, err := db.bucketConn.Get(key, &deviceInfo)
	if err != nil {
		return []Event{}, emperror.WrapWith(err, "Getting history from database failed", "device id", deviceId)
	}
	return deviceInfo, nil
}

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

func (db *DbConnection) RemoveHistory(deviceId string, numToRemove int) error {
	key := strings.Join([]string{historyDoc, deviceId}, ":")
	for a := 0; a < numToRemove; a++ {
		_, err := db.bucketConn.ListRemove(key, 0)
		return emperror.WrapWith(err, "Removing from history failed", "number of events successfully removed", a,
			"device id", deviceId)
	}
	return nil
}

func (db *DbConnection) InsertEvent(deviceId string, event Event, tombstoneMapKey string) error {
	if valid, err := isStateValid(deviceId, event); !valid {
		return emperror.WrapWith(err, "Insert event not attempted", "device id", deviceId,
			"event", event)
	}

	// get event id given device id
	counterKey := strings.Join([]string{counterDoc, deviceId}, ":")
	eventId, _, err := db.bucketConn.Counter(counterKey, 1, 0, 0)
	if err != nil {
		return emperror.WrapWith(err, "Failed to get event id", "device id", deviceId)
	}

	event.Id = strconv.FormatUint(eventId, 10)

	// append to the history, create if it doesn't exist (like that java thing?)
	historyKey := strings.Join([]string{historyDoc, deviceId}, ":")
	_, err = db.bucketConn.ListAppend(historyKey, &event, true)
	if err != nil {
		return emperror.WrapWith(err, "Failed to add event to history", "device id", deviceId,
			"event id", eventId, "event", event)
	}
	// update expiry time of the list document
	newTimeout := time.Now().Add(db.Timeout).Unix()
	_, err = db.bucketConn.Touch(historyKey, 0, uint32(newTimeout))
	if err != nil {
		return emperror.WrapWith(err, "Failed to update timeout", "device id", deviceId,
			"event id", eventId, "event", event)
	}

	//if tombstonekey isn't empty string, then set the tombstone map at that key
	if tombstoneMapKey != "" {
		tombstoneKey := strings.Join([]string{tombstoneDoc, deviceId}, ":")
		events := make(map[string]Event)
		events[tombstoneKey] = event
		_, err := db.bucketConn.Insert(tombstoneKey, &events, 0)
		if err != nil && err != gocb.ErrKeyExists {
			return emperror.WrapWith(err, "Failed to create tombstone", "device id", deviceId,
				"event id", eventId, "event", event)
		}
		_, err = db.bucketConn.MutateIn(tombstoneKey, 0, 0).
			Upsert(tombstoneMapKey, &event, false).
			Execute()
		if err != nil {
			return emperror.WrapWith(err, "Failed to add event to tombstone", "device id", deviceId,
				"event id", eventId, "event", event)
		}
	}

	return nil
}

func isStateValid(deviceId string, event Event) (bool, error) {
	if deviceId == "" {
		return false, errors.New("Invalid device id")
	}
	if event.Source == "" {
		return false, errors.New("Invalid event")
	}
	return true, nil
}

func (db *DbConnection) RemoveAll() error {
	_, err := db.bucketConn.ExecuteN1qlQuery(gocb.NewN1qlQuery("DELETE FROM devices;"), nil)
	if err != nil {
		return emperror.Wrap(err, "Removing all devices from database failed")
	}
	return nil
}
