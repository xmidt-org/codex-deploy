package main

import (
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/cassandra"
	"github.com/xmidt-org/webpa-common/xmetrics/xmetricstest"
)

func getConnection() (*cassandra.Connection, error) {
	return cassandra.CreateDbConnection(
		cassandra.Config{
			Hosts:    []string{"localhost:9042"},
			Database: "devices",
		},
		xmetricstest.NewProvider(nil),
		nil,
	)
}

func GetInserter() db.Inserter {
	conn, err := getConnection()
	if err != nil {
		panic(err)
	}
	return conn
}

func GetReader() db.RecordGetter {
	conn, err := getConnection()

	if err != nil {
		panic(err)
	}
	return conn
}
