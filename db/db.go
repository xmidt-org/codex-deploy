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

	"github.com/Comcast/codex/blacklist"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
)

var (
	errTableNotExist    = errors.New("Table does not exist")
	errInvaliddeviceID  = errors.New("Invalid device ID")
	errInvalidEventType = errors.New("Invalid event type")
	errNoEvents         = errors.New("no records to be inserted")
)

const (
	defaultPruneLimit     = 0
	defaultConnectTimeout = time.Duration(10) * time.Second
	defaultOpTimeout      = time.Duration(10) * time.Second
	defaultNumRetries     = 0
	defaultWaitTimeMult   = 1
	defaultPingInterval   = time.Second
	defaultMaxIdleConns   = 2
	defaultMaxOpenConns   = 0
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
	PruneLimit     int
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
	findList    findList
	mutliInsert multiinserter
	deleter     deleter
	closer      closer
	pinger      pinger
	stats       stats
	gennericDB  *sql.DB

	pruneLimit  int
	health      *health.Health
	measures    Measures
	stopThreads []chan struct{}
}

type Record struct {
	Type      EventType `json:"type" gorm:"type:int"`
	DeviceID  string    `json:"deviceid"`
	BirthDate int64     `json:"birthdate"`
	DeathDate int64     `json:"deathdate"`
	Data      []byte    `json:"data"`
	Nonce     []byte    `json:"nonce"`
	Alg       string    `json:"alg"`
	KID       string    `json:"kid" gorm:"Column:kid"`
}

// set Record's table name to be `events`
func (Record) TableName() string {
	return "events"
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(config Config, provider provider.Provider, health *health.Health) (*Connection, error) {
	var (
		conn          *dbDecorator
		err           error
		connectionURL string
	)

	db := Connection{
		health:     health,
		pruneLimit: config.PruneLimit,
	}

	validateConfig(&config)

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
			config.Database + "?sslmode=verify-full&sslrootcert=" + config.SSLRootCert +
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

	emptyRecord := Record{}
	if !conn.HasTable(&emptyRecord) {
		return &Connection{}, emperror.WrapWith(errTableNotExist, "Connecting to database failed", "table name", emptyRecord.TableName())
	}

	db.finder = conn
	db.findList = conn
	db.mutliInsert = conn
	db.deleter = conn
	db.closer = conn
	db.pinger = conn
	db.stats = conn
	db.gennericDB = conn.DB.DB()
	db.measures = NewMeasures(provider)

	db.setupHealthCheck(config.PingInterval)
	db.setupMetrics()
	db.configure(config.MaxIdleConns, config.MaxOpenConns)

	return &db, nil
}

func validateConfig(config *Config) {
	zeroDuration := time.Duration(0) * time.Second

	// TODO: check if username, server, or database is empty?

	if config.PruneLimit < 0 {
		config.PruneLimit = defaultPruneLimit
	}
	if config.ConnectTimeout == zeroDuration {
		config.ConnectTimeout = defaultConnectTimeout
	}
	if config.OpTimeout == zeroDuration {
		config.OpTimeout = defaultOpTimeout
	}
	if config.NumRetries < 0 {
		config.NumRetries = defaultNumRetries
	}
	if config.WaitTimeMult < 1 {
		config.WaitTimeMult = defaultWaitTimeMult
	}
	if config.PingInterval == zeroDuration {
		config.PingInterval = defaultPingInterval
	}
	if config.MaxIdleConns < 2 {
		config.MaxIdleConns = defaultMaxIdleConns
	}
	if config.MaxOpenConns < 0 {
		config.MaxOpenConns = defaultMaxOpenConns
	}
}

func (db *Connection) configure(maxIdleConns int, maxOpenConns int) {
	db.gennericDB.SetMaxIdleConns(maxIdleConns)
	db.gennericDB.SetMaxOpenConns(maxOpenConns)
}

func (db *Connection) setupHealthCheck(interval time.Duration) {
	if db.health == nil {
		return
	}
	sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
		Pinger: db.gennericDB,
	})
	if err != nil {
		// todo: capture this error somehow
	}

	db.health.AddCheck(&health.Config{
		Name:     "sql-check",
		Checker:  sqlCheck,
		Interval: interval,
		Fatal:    true,
	})
}

func (db *Connection) setupMetrics() {
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
func (db *Connection) GetRecords(deviceID string, limit int) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	err := db.finder.findRecords(&deviceInfo, limit, "device_id = ?", deviceID)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, readType).Add(1.0)
		return []Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	db.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	db.measures.SQLQuerySuccessCount.With(typeLabel, readType).Add(1.0)
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device
func (db *Connection) GetRecordsOfType(deviceID string, limit int, eventType EventType) ([]Record, error) {
	var (
		deviceInfo []Record
	)
	err := db.finder.findRecords(&deviceInfo, limit, "device_id = ? AND type = ?", deviceID, eventType)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, readType).Add(1.0)
		return []Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	db.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	db.measures.SQLQuerySuccessCount.With(typeLabel, readType).Add(1.0)
	return deviceInfo, nil
}

// GetBlacklist returns a list of blacklisted devices
func (db *Connection) GetBlacklist() (list []blacklist.BlackListedItem, err error) {
	err = db.findList.findBlacklist(&list)
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, listReadType).Add(1.0)
		return []blacklist.BlackListedItem{}, emperror.WrapWith(err, "Getting records from database failed")
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, listReadType).Add(1.0)
	return
}

// PruneRecords removes records past their deathdate.
func (db *Connection) PruneRecords(t int64) error {
	rowsAffected, err := db.deleter.delete(&Record{}, db.pruneLimit, "death_date < ?", t)
	db.measures.SQLDeletedRecords.Add(float64(rowsAffected))
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, deleteType).Add(1.0)
		return emperror.WrapWith(err, "Prune records failed", "time", t)
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, deleteType).Add(1.0)
	return nil
}

// InsertEvent adds a record to the table.
func (db *Connection) InsertRecords(records ...Record) error {
	rowsAffected, err := db.mutliInsert.insert(records)
	db.measures.SQLInsertedRecords.Add(float64(rowsAffected))
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, insertType).Add(1.0)
		return emperror.Wrap(err, "Inserting records failed")
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
	rowsAffected, err := db.deleter.delete(&Record{}, 0)
	db.measures.SQLDeletedRecords.Add(float64(rowsAffected))
	if err != nil {
		db.measures.SQLQueryFailureCount.With(typeLabel, deleteType).Add(1.0)
		return emperror.Wrap(err, "Removing all records from database failed")
	}
	db.measures.SQLQuerySuccessCount.With(typeLabel, deleteType).Add(1.0)
	return nil
}
