package mongodb

import (
	"testing"

	"github.com/Comcast/codex/db"
	"github.com/stretchr/testify/assert"
)

func TestImplementsInterfaces(t *testing.T) {
	var (
		dbConn interface{}
	)
	assert := assert.New(t)
	dbConn = &Connection{}
	_, ok := dbConn.(db.Inserter)
	assert.True(ok)
	_, ok = dbConn.(db.RecordGetter)
	assert.True(ok)
}
