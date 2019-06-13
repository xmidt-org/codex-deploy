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

// set Record's table name to be `events`
func (Record) TableName() string {
	return "events"
}

type Inserter interface {
	InsertRecords(records ...Record) error
}

type Pruner interface {
	GetRecordIDs(shard int, limit int, deathDate int64) ([]int, error)
	PruneRecords(records []int) error
}

type RecordGetter interface {
	GetRecords(deviceID string, limit int) ([]Record, error)
	GetRecordsOfType(deviceID string, limit int, eventType EventType) ([]Record, error)
}
