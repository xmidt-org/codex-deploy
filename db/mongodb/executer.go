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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	// Import GORM-related packages.

	"context"

	"github.com/Comcast/codex/blacklist"
	"github.com/Comcast/codex/db"
	"github.com/goph/emperror"
	"go.mongodb.org/mongo-driver/mongo"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type (
	finder interface {
		findRecords(ctx context.Context, limit int, filter interface{}) ([]db.Record, error)
	}
	findList interface {
		findBlacklist(ctx context.Context) ([]blacklist.BlackListedItem, error)
	}
	multiinserter interface {
		insert(ctx context.Context, records []db.Record) (int, error)
	}
	pinger interface {
		ping(ctx context.Context) error
	}
	closer interface {
		close(ctx context.Context) error
	}
)

type dbDecorator struct {
	client    *mongo.Client
	records   *mongo.Collection
	blacklist *mongo.Collection
}

func (b *dbDecorator) findRecords(ctx context.Context, limit int, filter interface{}) ([]db.Record, error) {
	var (
		records []db.Record
	)
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{})
	cursor, err := b.records.Find(ctx, filter, opts)
	if err != nil {
		return []db.Record{}, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var record db.Record
		err := cursor.Decode(&record)
		if err != nil {
			return []db.Record{}, emperror.Wrap(err, "failed to decode a record")
		}
		records = append(records, record)

	}
	if err := cursor.Err(); err != nil {
		return []db.Record{}, err
	}
	return records, nil
}

func (b *dbDecorator) findBlacklist(ctx context.Context) ([]blacklist.BlackListedItem, error) {
	var (
		items []blacklist.BlackListedItem
	)
	opts := options.Find().SetLimit(0)
	cursor, err := b.blacklist.Find(ctx, bson.D{}, opts)
	if err != nil {
		return []blacklist.BlackListedItem{}, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var item blacklist.BlackListedItem
		err := cursor.Decode(&item)
		if err != nil {
			return []blacklist.BlackListedItem{}, emperror.Wrap(err, "failed to decode a record")
		}
		items = append(items, item)

	}
	if err := cursor.Err(); err != nil {
		return []blacklist.BlackListedItem{}, err
	}
	return items, nil

}

func (b *dbDecorator) insert(ctx context.Context, records []db.Record) (int, error) {
	var (
		thingsToInsert []interface{}
	)
	for _, record := range records {
		thingsToInsert = append(thingsToInsert, record)
	}
	result, err := b.records.InsertMany(ctx, thingsToInsert)
	return len(result.InsertedIDs), err
}

func (b *dbDecorator) ping(ctx context.Context) error {
	return b.client.Ping(ctx, nil)
}

func (b *dbDecorator) close(ctx context.Context) error {
	return b.client.Disconnect(ctx)
}

func connect(ctx context.Context, uri string, dbName string) (*dbDecorator, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	database := client.Database(dbName)
	records := database.Collection("events")
	db := &dbDecorator{
		client:    client,
		blacklist: database.Collection("blacklist"),
		records:   records,
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	modelOpts := options.Index().SetExpireAfterSeconds(0)
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "deathdate", Value: 1}},
		Options: modelOpts,
	}
	records.Indexes().CreateOne(ctx, model, opts)
	return db, nil
}
