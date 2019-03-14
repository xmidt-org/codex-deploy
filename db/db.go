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
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"
)

var (
	errInvaliddeviceID  = errors.New("Invalid device ID")
	errInvalidEventType = errors.New("Invalid event type")
	errNoEvents         = errors.New("no records to be inserted")
)

const (
	typeLabel  = "type"
	insertType = "insert"
	deleteType = "delete"
	readType   = "read"
	pingType   = "ping"
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
	ConnectTimeout time.Duration
	OpTimeout      time.Duration

	// MaxIdleConns sets the max idle connections, the min value is 2
	MaxIdleConns int

	// MaxOpenConns sets the max open connections, to specify unlimited set to 0
	MaxOpenConns int

	PingInterval time.Duration
}

// Connection contains the tools to edit the database.
type Connection struct {
	finder      finder
	mutliInsert multiinserter
	deleter     deleter
	closer      closer
	pinger      pinger
	stats       stats
	gennericDB  *sql.DB

	measures    Measures
	stopThreads []chan struct{}
}

// Event represents the event information in the database.  It has no TTL.
//
// swagger:model Event
type Event struct {
	// The id for the event.
	//
	// required: true
	// example: 425808997514969090
	ID int `json:"id"`

	// The time this event was found.
	//
	// required: true
	// example: 1549969802
	Time int64 `json:"time"`

	// The source of this event.
	//
	// required: true
	// example: dns:talaria-1234
	Source string `json:"src"`

	// The destination of this event.
	//
	// required: true
	// example: device-status/5/offline
	Destination string `json:"dest"`

	// The partners related to this device.
	//
	// required: true
	// example: ["hello","world"]
	PartnerIDs []string `json:"partner_ids"`

	// The transaction id for this event.
	//
	// required: true
	// example: AgICJpZCI6ICJtYWM6NDhmN2MwZDc5MDI0Iiw
	TransactionUUID string `json:"transaction_uuid,omitempty"`

	// list of bytes received from the source.
	// If the device destination matches "device-status/.*", this is a base64
	// encoded json map that contains the key "ts", denoting the time the event
	// was created.
	//
	// required: false
	// example: eyJpZCI6IjUiLCJ0cyI6IjIwMTktMDItMTJUMTE6MTA6MDIuNjE0MTkxNzM1WiIsImJ5dGVzLXNlbnQiOjAsIm1lc3NhZ2VzLXNlbnQiOjEsImJ5dGVzLXJlY2VpdmVkIjowLCJtZXNzYWdlcy1yZWNlaXZlZCI6MH0=
	Payload []byte `json:"payload,omitempty"`

	// Other metadata and details related to this state.
	//
	// required: true
	// example: {"/boot-time":1542834188,"/last-reconnect-reason":"spanish inquisition"}
	Details map[string]interface{} `json:"details"`
}

type Record struct {
	ID        int       `json:"id" gorm:"AUTO_INCREMENT"`
	Type      int       `json:"type"`
	DeviceID  string    `json:"deviceid" gorm:"not null;index"`
	BirthDate time.Time `json:"birthdate" gorm:"index"`
	DeathDate time.Time `json:"deathdate" gorm:"index"`
	Data      []byte    `json:"data" gorm:"not null"`
}

// set Record's table name to be `events`
func (Record) TableName() string {
	return "events"
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(config Config, provider provider.Provider) (*Connection, error) {
	var (
		conn          *dbDecorator
		err           error
		connectionURL string
	)

	db := Connection{}

	// pq expects seconds
	connectTimeout := strconv.Itoa(int(config.ConnectTimeout.Seconds()))

	// pq expects milliseconds
	opTimeout := strconv.Itoa(int(float64(config.OpTimeout.Nanoseconds()) / 1000000))

	// include timeout when connecting
	// if missing a cert, connect insecurely
	if config.SSLCert == "" || config.SSLKey == "" || config.SSLRootCert == "" {
		connectionURL = "postgresql://" + config.Username + "@" + config.Server + "/" +
			config.Database + "?sslmode=disable&connect_timeout=" + connectTimeout +
			"&statement_timeout=" + opTimeout
	} else {
		connectionURL = "postgresql://" + config.Username + "@" + config.Server + "/" +
			config.Database + "?ssl=true&sslmode=verify-full&sslrootcert=" + config.SSLRootCert +
			"&sslkey=" + config.SSLKey + "&sslcert=" + config.SSLCert + "&connect_timeout=" +
			connectTimeout + "&statement_timeout=" + opTimeout
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
		return &Connection{}, emperror.WrapWith(err, "Connecting to database failed", "connection url", connectionURL)
	}

	conn.AutoMigrate(&Record{})

	db.finder = conn
	db.mutliInsert = conn
	db.deleter = conn
	db.closer = conn
	db.pinger = conn
	db.stats = conn
	db.gennericDB = conn.DB.DB()
	db.measures = NewMeasures(provider)

	db.setupMetrics()
	db.configure(config.MaxIdleConns, config.MaxOpenConns)

	return &db, nil
}

func (db *Connection) configure(maxIdleConns int, maxOpenConns int) {
	if maxIdleConns < 2 {
		maxIdleConns = 2
	}
	db.gennericDB.SetMaxIdleConns(maxIdleConns)
	db.gennericDB.SetMaxOpenConns(maxOpenConns)
}

func (db *Connection) setupMetrics() {
	// ping to check status
	pingStop := doEvery(time.Second, func() {
		err := db.Ping()
		if err != nil {
			db.measures.ConnectionStatus.Set(0.0)
		} else {
			db.measures.ConnectionStatus.Set(1.0)
		}
	})
	db.stopThreads = append(db.stopThreads, pingStop)

	// baseline
	startStats := db.stats.getStats()
	prevWaitCount := startStats.WaitCount
	prevWaitDuration := startStats.WaitDuration.Nanoseconds()
	prevMaxIdleClosed := startStats.MaxIdleClosed
	prevMaxLifetimeClosed := startStats.MaxLifetimeClosed

	// update measurements
	metricsStop := doEvery(time.Second, func() {
		stats := db.stats.getStats()

		// current connections
		db.measures.PoolOpenConnections.Set(float64(stats.OpenConnections))
		db.measures.PoolInUseConnections.Set(float64(stats.InUse))
		db.measures.PoolIdleConnections.Set(float64(stats.Idle))

		// Counters
		db.measures.SQLWaitCount.Add(float64(stats.WaitCount - prevWaitCount))
		db.measures.SQLWaitDuration.Add(float64(stats.WaitDuration.Nanoseconds() - prevWaitDuration))
		db.measures.SQLMaxIdleClosed.Add(float64(stats.MaxIdleClosed - prevMaxIdleClosed))
		db.measures.SQLMaxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed - prevMaxLifetimeClosed))
	})
	db.stopThreads = append(db.stopThreads, metricsStop)
}

// GetRecords returns a list of records for a given device
func (db *Connection) GetRecords(deviceID string) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	err := db.finder.find(&deviceInfo, "device_id = ?", deviceID)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, readType).Add(1.0)
		return []Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, readType).Add(1.0)
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device
func (db *Connection) GetRecordsOfType(deviceID string, eventType int) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	err := db.finder.find(&deviceInfo, "device_id = ? AND type = ?", deviceID, eventType)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, readType).Add(1.0)
		return []Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, readType).Add(1.0)
	return deviceInfo, nil
}

// PruneRecords removes records past their deathdate.
func (db *Connection) PruneRecords(t time.Time) error {
	rowsAffected, err := db.deleter.delete(&Record{}, "death_date < ?", t)
	db.measures.SQLDeletedRows.Add(float64(rowsAffected))
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, deleteType).Add(1.0)
		return emperror.WrapWith(err, "Prune records failed", "time", t)
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, deleteType).Add(1.0)
	return nil
}

// InsertEvent adds a record to the table.
func (db *Connection) InsertRecords(records ...Record) error {
	err := db.mutliInsert.insert(records)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, insertType).Add(1.0)
		return emperror.WrapWith(err, "Inserting records failed", "records", records)
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, insertType).Add(1.0)
	return nil
}

func (db *Connection) Ping() error {
	err := db.pinger.ping()
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, pingType).Add(1.0)
		return emperror.WrapWith(err, "Pinging connection failed")
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, pingType).Add(1.0)
	return nil
}

func (db *Connection) Close() error {
	for _, stopThread := range db.stopThreads {
		stopThread <- struct{}{}
	}

	err := db.closer.close()
	if err != nil {
		return emperror.WrapWith(err, "Closing connection failed")
	}
	return nil
}

func doEvery(d time.Duration, f func()) chan struct{} {
	ticker := time.NewTicker(d)
	stop := make(chan struct{}, 1)
	go func(stop chan struct{}) {
		for {
			select {
			case <-ticker.C:
				f()
			case <-stop:
				return
			}
		}
	}(stop)
	return stop
}

// RemoveAll removes everything in the events table.  Used for testing.
func (db *Connection) RemoveAll() error {
	rowsAffected, err := db.deleter.delete(&Record{})
	db.measures.SQLDeletedRows.Add(float64(rowsAffected))
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, deleteType).Add(1.0)
		return emperror.Wrap(err, "Removing all records from database failed")
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, deleteType).Add(1.0)
	return nil
}
