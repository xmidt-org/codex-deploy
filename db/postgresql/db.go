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

// package postgresql provides a way to connect to a postgresql database to
// keep track of device events.
package postgresql

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/Comcast/codex/blacklist"
	"github.com/Comcast/codex/db"

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

// Config contains the initial configuration information needed to create a
// postgresql db connection.
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

// Connection manages the connection to the postgresql database, and maintains
// a health check on the database connection.
type Connection struct {
	finder       finder
	findList     findList
	deviceFinder deviceFinder
	mutliInsert  multiinserter
	deleter      deleter
	closer       closer
	pinger       pinger
	stats        stats
	gennericDB   *sql.DB

	pruneLimit  int
	health      *health.Health
	measures    Measures
	stopThreads []chan struct{}
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(config Config, provider provider.Provider, health *health.Health) (*Connection, error) {
	var (
		conn          *dbDecorator
		err           error
		connectionURL string
	)

	dbConn := Connection{
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

	emptyRecord := db.Record{}
	if !conn.HasTable(&emptyRecord) {
		return &Connection{}, emperror.WrapWith(errTableNotExist, "Connecting to database failed", "table name", emptyRecord.TableName())
	}

	dbConn.finder = conn
	dbConn.findList = conn
	dbConn.deviceFinder = conn
	dbConn.mutliInsert = conn
	dbConn.deleter = conn
	dbConn.closer = conn
	dbConn.pinger = conn
	dbConn.stats = conn
	dbConn.gennericDB = conn.DB.DB()
	dbConn.measures = NewMeasures(provider)

	dbConn.setupHealthCheck(config.PingInterval)
	dbConn.setupMetrics()
	dbConn.configure(config.MaxIdleConns, config.MaxOpenConns)

	return &dbConn, nil
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

func (c *Connection) configure(maxIdleConns int, maxOpenConns int) {
	c.gennericDB.SetMaxIdleConns(maxIdleConns)
	c.gennericDB.SetMaxOpenConns(maxOpenConns)
}

func (c *Connection) setupHealthCheck(interval time.Duration) {
	if c.health == nil {
		return
	}
	sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
		Pinger: c.gennericDB,
	})
	if err != nil {
		// todo: capture this error somehow
	}

	c.health.AddCheck(&health.Config{
		Name:     "sql-check",
		Checker:  sqlCheck,
		Interval: interval,
		Fatal:    true,
	})
}

func (c *Connection) setupMetrics() {
	// baseline
	startStats := c.stats.getStats()
	prevWaitCount := startStats.WaitCount
	prevWaitDuration := startStats.WaitDuration.Nanoseconds()
	prevMaxIdleClosed := startStats.MaxIdleClosed
	prevMaxLifetimeClosed := startStats.MaxLifetimeClosed

	// update measurements
	metricsStop := doEvery(time.Second, func() {
		stats := c.stats.getStats()

		// current connections
		c.measures.PoolOpenConnections.Set(float64(stats.OpenConnections))
		c.measures.PoolInUseConnections.Set(float64(stats.InUse))
		c.measures.PoolIdleConnections.Set(float64(stats.Idle))

		// Counters
		c.measures.SQLWaitCount.Add(float64(stats.WaitCount - prevWaitCount))
		c.measures.SQLWaitDuration.Add(float64(stats.WaitDuration.Nanoseconds() - prevWaitDuration))
		c.measures.SQLMaxIdleClosed.Add(float64(stats.MaxIdleClosed - prevMaxIdleClosed))
		c.measures.SQLMaxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed - prevMaxLifetimeClosed))
	})
	c.stopThreads = append(c.stopThreads, metricsStop)
}

// GetRecords returns a list of records for a given device.
func (c *Connection) GetRecords(deviceID string, limit int) ([]db.Record, error) {
	var (
		deviceInfo []db.Record
	)
	err := c.finder.findRecords(&deviceInfo, limit, "device_id = ?", deviceID)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device and event type.
func (c *Connection) GetRecordsOfType(deviceID string, limit int, eventType db.EventType) ([]db.Record, error) {
	var (
		deviceInfo []db.Record
	)
	err := c.finder.findRecords(&deviceInfo, limit, "device_id = ? AND type = ?", deviceID, eventType)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetRecordsToDelete returns a list of record ids and deathdates not past a
// given date.
func (c *Connection) GetRecordsToDelete(shard int, limit int, deathDate int64) ([]db.RecordToDelete, error) {
	recordsToDelete, err := c.finder.findRecordsToDelete(limit, shard, deathDate)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.RecordToDelete{}, emperror.WrapWith(err, "Getting record IDs from database failed", "shard", shard, "death date", deathDate)
	}
	c.measures.SQLReadRecords.Add(float64(len(recordsToDelete)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return recordsToDelete, nil
}

// GetBlacklist returns a list of blacklisted devices.
func (c *Connection) GetBlacklist() (list []blacklist.BlackListedItem, err error) {
	err = c.findList.findBlacklist(&list)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.BlacklistReadType).Add(1.0)
		return []blacklist.BlackListedItem{}, emperror.WrapWith(err, "Getting records from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.BlacklistReadType).Add(1.0)
	return
}

// GetDeviceList returns a list of device ids where the device id is greater
// than the offset device id.
func (c *Connection) GetDeviceList(offset string, limit int) ([]string, error) {
	list, err := c.deviceFinder.getList(offset, limit)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []string{}, emperror.WrapWith(err, "Getting list of devices from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return list, nil
}

// DeleteRecord removes a record.
func (c *Connection) DeleteRecord(shard int, deathDate int64, recordID int64) error {
	rowsAffected, err := c.deleter.delete(&db.Record{}, 1, "shard = ? AND death_date = ? AND record_id = ?", shard, deathDate, recordID)
	c.measures.SQLDeletedRecords.Add(float64(rowsAffected))
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
		return emperror.WrapWith(err, "Prune records failed", "record id", recordID)
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
	return nil
}

// InsertEvent adds a list of records to the table.
func (c *Connection) InsertRecords(records ...db.Record) error {
	rowsAffected, err := c.mutliInsert.insert(records)
	c.measures.SQLInsertedRecords.Add(float64(rowsAffected))
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.InsertType).Add(1.0)
		return emperror.Wrap(err, "Inserting records failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.InsertType).Add(1.0)
	return nil
}

// Ping is for pinging the database to verify that the connection is still good.
func (c *Connection) Ping() error {
	err := c.pinger.ping()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.PingType).Add(1.0)
		return emperror.WrapWith(err, "Pinging connection failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.PingType).Add(1.0)
	return nil
}

// Close closes the database connection.
func (c *Connection) Close() error {
	for _, stopThread := range c.stopThreads {
		stopThread <- struct{}{}
	}

	err := c.closer.close()
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
func (c *Connection) RemoveAll() error {
	rowsAffected, err := c.deleter.delete(&db.Record{}, 0)
	c.measures.SQLDeletedRecords.Add(float64(rowsAffected))
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
		return emperror.Wrap(err, "Removing all records from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.DeleteType).Add(1.0)
	return nil
}
