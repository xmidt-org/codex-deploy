package db

const (
	TypeLabel    = "type"
	InsertType   = "insert"
	DeleteType   = "delete"
	ReadType     = "read"
	PingType     = "ping"
	ListReadType = "listRead"
)

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

type RecordToDelete struct {
	DeathDate int64 `json:"deathdate" bson:"deathdate"`
	RecordID  int64 `json:"recordid" bson:"recordid"`
}

// set Record's table name to be `events`
func (Record) TableName() string {
	return "events"
}

type Inserter interface {
	InsertRecords(records ...Record) error
}

type Pruner interface {
	GetRecordsToDelete(shard int, limit int, deathDate int64) ([]RecordToDelete, error)
	// PruneRecords(records []int) error
	DeleteRecord(shard int, deathdate int64, recordID int64) error
}

type RecordGetter interface {
	GetRecords(deviceID string, limit int) ([]Record, error)
	GetRecordsOfType(deviceID string, limit int, eventType EventType) ([]Record, error)
}
