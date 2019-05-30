package batchInserter

import (
	"github.com/Comcast/codex/db"
	"github.com/stretchr/testify/mock"
)

type mockInserter struct {
	mock.Mock
}

func (c *mockInserter) InsertRecords(records ...db.Record) error {
	args := c.Called(records)
	return args.Error(0)
}
