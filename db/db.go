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
	"time"

	"github.com/goph/emperror"
)

// Interface describes the main functionality needed to connect to a database.
type Interface interface {
	GetRecords(deviceID string) ([]Record, error)
	GetRecordsOfType(deviceID string, eventType int) ([]Record, error)
	PruneRecords(t time.Time) error
	InsertRecord(record Record) error
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

// Config contains the initial configuration information needed to create a db connection.
type Config struct {
	Server         string
	Username       string
	Database       string
	SSLRootCert    string
	SSLKey         string
	SSLCert        string
	NumRetries     int
	WaitTimeMult   time.Duration
	ConnectTimeout string
	OpTimeout      string
}

// Connection contains the tools to edit the database.
type Connection struct {
	finder  finder
	creator creator
	deleter deleter
}

// Event represents the event information in the database.  It has no TTL.
//
// swagger:model Event
type Event struct {
	// The id for the event.
	//
	// required: true
	ID int `json:"id"`

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

type Record struct {
	ID        int       `json:"id" gorm:"AUTO_INCREMENT"`
	Type      int       `json:"type" gorm:"index"`
	DeviceID  string    `json:"deviceid" gorm:"not null;index"`
	BirthDate time.Time `json:"birthdate" gorm:"index"`
	DeathDate time.Time `json:"deathdate" gorm:"index"`
	Data      []byte    `json:"data" gorm:"not null"`
}

// set User's table name to be `profiles`
func (Record) TableName() string {
	return "events"
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(config Config) (*Connection, error) {
	var (
		conn          *dbDecorator
		err           error
		connectionURL string
	)

	db := Connection{}

	// include timeout when connecting
	// if missing a cert, connect insecurely
	if config.SSLCert == "" || config.SSLKey == "" || config.SSLRootCert == "" {
		connectionURL = "postgresql://" + config.Username + "@" + config.Server + "/" +
			config.Database + "?sslmode=disable&connect_timeout=" + config.ConnectTimeout +
			"&statement_timeout=" + config.OpTimeout
	} else {
		connectionURL = "postgresql://" + config.Username + "@" + config.Server + "/" +
			config.Database + "?ssl=true&sslmode=verify-full&sslrootcert=" + config.SSLRootCert +
			"&sslkey=" + config.SSLKey + "&sslcert=" + config.SSLCert + "&connect_timeout=" +
			config.ConnectTimeout + "&statement_timeout=" + config.OpTimeout
	}

	conn, err = connect(connectionURL)

	// retry if it fails
	waitTime := 1 * time.Second
	for attempt := 0; attempt < config.NumRetries && err != nil; attempt++ {
		time.Sleep(waitTime)
		conn, err = connect(connectionURL)
		waitTime = waitTime * config.WaitTimeMult
	}

	if err != nil {
		return &Connection{}, emperror.WrapWith(err, "Connecting to couchbase failed", "connection url", connectionURL)
	}

	conn.AutoMigrate(&Record{})

	db.finder = conn
	db.creator = conn
	db.deleter = conn

	return &db, nil
}

// GetRecords returns a list of records for a given device
func (db *Connection) GetRecords(deviceID string) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	if deviceID == "" {
		return []Record{}, emperror.WrapWith(errInvaliddeviceID, "Get tombstone not attempted",
			"device id", deviceID)
	}
	err := db.finder.find(&deviceInfo, "device_id = ?", deviceID)
	if err != nil {
		return []Record{}, emperror.WrapWith(err, "Getting tombstone from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device
func (db *Connection) GetRecordsOfType(deviceID string, eventType int) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	if deviceID == "" {
		return []Record{}, emperror.WrapWith(errInvaliddeviceID, "Get tombstone not attempted",
			"device id", deviceID)
	}
	err := db.finder.find(&deviceInfo, "device_id = ? AND type = ?", deviceID, eventType)
	if err != nil {
		return []Record{}, emperror.WrapWith(err, "Getting tombstone from database failed", "device id", deviceID)
	}
	return deviceInfo, nil
}

// PruneRecords removes records past their deathdate.
func (db *Connection) PruneRecords(t time.Time) error {
	err := db.deleter.delete(&Record{}, "death_date < ?", t)
	if err != nil {
		return emperror.WrapWith(err, "Prune events failed", "time", t)
	}
	return nil
}

// InsertEvent adds a record to the table.
func (db *Connection) InsertRecord(record Record) error {
	if valid, err := isRecordValid(record); !valid {
		return emperror.WrapWith(err, "Insert event not attempted", "record", record)
	}
	err := db.creator.create(&record)
	if err != nil {
		return emperror.WrapWith(err, "inserting event failed", "record", record)
	}
	return err
}

func isRecordValid(record Record) (bool, error) {
	if record.DeviceID == "" {
		return false, errInvaliddeviceID
	}
	return true, nil
}

// RemoveAll removes everything in the events table.  Used for testing.
func (db *Connection) RemoveAll() error {
	err := db.deleter.delete(&Record{})
	if err != nil {
		return emperror.Wrap(err, "Removing all devices from database failed")
	}
	return nil
}
