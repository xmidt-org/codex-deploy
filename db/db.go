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

// package db provides some overarching variables, structs, and interfaces
// that different database implementations can use and implement and consumers
// can expect.
package db

const (
	// TypeLabel is for labeling metrics; if there is a single metric for
	// successful queries, the typeLabel and corresponding type can be used
	// when incrementing the metric.
	TypeLabel  = "type"
	InsertType = "insert"
	DeleteType = "delete"
	ReadType   = "read"
	PingType   = "ping"
	// ListReadType is for reading from the blacklist.
	ListReadType = "listRead"
)

// Record is the struct used to insert an event into the database.  It includes
// the marshaled, and possibly encrypted, event ("Data") and then other
// metadata to be used for the record.  If the data is encrypted, the Nonce,
// Alg, and KID values will be needed to determine how to correctly decrypt it.
type Record struct {
	Type      EventType `json:"type" bson:"type" gorm:"type:int"`
	DeviceID  string    `json:"deviceid" bson:"deviceid"`
	BirthDate int64     `json:"birthdate" bson:"birthdate"`
	DeathDate int64     `json:"deathdate" bson:"deathdate"`
	Data      []byte    `json:"data" bson:"data"`
	Nonce     []byte    `json:"nonce" bson:"nonce"`
	Alg       string    `json:"alg" bson:"alg"`
	KID       string    `json:"kid" bson:"kid" gorm:"Column:kid"`
}

// RecordToDelete is the information needed to get out of the database in order
// to call the DeleteRecord function
type RecordToDelete struct {
	DeathDate int64 `json:"deathdate" bson:"deathdate"`
	RecordID  int64 `json:"recordid" bson:"recordid"`
}

// TableName sets Record's table name to be "events"; for the GORM driver.
func (Record) TableName() string {
	return "events"
}

// Inserter is something that can insert records into the database.
type Inserter interface {
	InsertRecords(records ...Record) error
}

// Pruner is something that can get a list of expired records and delete them.
// Deleting is done individually.
type Pruner interface {
	GetRecordsToDelete(shard int, limit int, deathDate int64) ([]RecordToDelete, error)
	// PruneRecords(records []int) error
	DeleteRecord(shard int, deathdate int64, recordID int64) error
}

// RecordGetter is something that can get records, including only getting records of a
// certain type.
type RecordGetter interface {
	GetRecords(deviceID string, limit int) ([]Record, error)
	GetRecordsOfType(deviceID string, limit int, eventType EventType) ([]Record, error)
}
