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

package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/Comcast/codex/blacklist"
	"github.com/Comcast/codex/db"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"

	"github.com/InVisionApp/go-health"
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
	Password       string
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
	closer      closer
	pinger      pinger

	opTimeout   time.Duration
	pruneLimit  int
	health      *health.Health
	measures    Measures
	stopThreads []chan struct{}
}

// CreateDbConnection creates db connection and returns the struct to the consumer.
func CreateDbConnection(config Config, provider provider.Provider, health *health.Health) (*Connection, error) {
	var (
		conn *dbDecorator
		err  error
	)

	dbConn := Connection{
		health:     health,
		pruneLimit: config.PruneLimit,
	}

	validateConfig(&config)

	connectionURL := "mongodb://" + config.Username + ":" + config.Password + "@" + config.Server

	if config.Username == "" || config.Password == "" {
		connectionURL = "mongodb://" + config.Server
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	conn, err = connect(ctx, connectionURL, config.Database)
	cancel()

	// retry if it fails
	waitTime := 1 * time.Second
	for attempt := 0; attempt < config.NumRetries && err != nil; attempt++ {
		time.Sleep(waitTime)
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		conn, err = connect(ctx, connectionURL, config.Database)
		cancel()
		waitTime = waitTime * config.WaitTimeMult
	}

	if err != nil {
		return &Connection{}, emperror.WrapWith(err, "Connecting to database failed", "connection url", connectionURL)
	}

	dbConn.opTimeout = config.OpTimeout

	dbConn.finder = conn
	dbConn.findList = conn
	dbConn.mutliInsert = conn
	dbConn.closer = conn
	dbConn.pinger = conn
	dbConn.measures = NewMeasures(provider)

	dbConn.setupHealthCheck(config.PingInterval)

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

func (c *Connection) setupHealthCheck(interval time.Duration) {
	// if c.health == nil {
	// 	return
	// }
	// sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
	// 	Pinger: c.pinger,
	// })
	// if err != nil {
	// 	// todo: capture this error somehow
	// }

	// c.health.AddCheck(&health.Config{
	// 	Name:     "sql-check",
	// 	Checker:  sqlCheck,
	// 	Interval: interval,
	// 	Fatal:    true,
	// })
}

// GetRecords returns a list of records for a given device
func (c *Connection) GetRecords(deviceID string, limit int) ([]db.Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	deviceInfo, err := c.finder.findRecords(ctx, limit, bson.D{{Key: "deviceid", Value: deviceID}})
	cancel()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device
func (c *Connection) GetRecordsOfType(deviceID string, limit int, eventType db.EventType) ([]db.Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	deviceInfo, err := c.finder.findRecords(ctx, limit, bson.D{{Key: "deviceid", Value: deviceID}, {Key: "type", Value: eventType}})
	cancel()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetBlacklist returns a list of blacklisted devices
func (c *Connection) GetBlacklist() (list []blacklist.BlackListedItem, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	list, err = c.findList.findBlacklist(ctx)
	cancel()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ListReadType).Add(1.0)
		return []blacklist.BlackListedItem{}, emperror.WrapWith(err, "Getting records from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ListReadType).Add(1.0)
	return
}

// InsertEvent adds a record to the table.
func (c *Connection) InsertRecords(records ...db.Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	rowsAffected, err := c.mutliInsert.insert(ctx, records)
	cancel()
	c.measures.SQLInsertedRecords.Add(float64(rowsAffected))
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.InsertType).Add(1.0)
		return emperror.Wrap(err, "Inserting records failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.InsertType).Add(1.0)
	return nil
}

func (c *Connection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	err := c.pinger.ping(ctx)
	cancel()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.PingType).Add(1.0)
		return emperror.WrapWith(err, "Pinging connection failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.PingType).Add(1.0)
	return nil
}

func (c *Connection) Close() error {
	for _, stopThread := range c.stopThreads {
		stopThread <- struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.opTimeout)
	err := c.closer.close(ctx)
	cancel()
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
