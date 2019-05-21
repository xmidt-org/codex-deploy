package batchInserter

import (
	"errors"
	"sync"
	"time"

	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/logging"
	"github.com/Comcast/webpa-common/semaphore"
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/go-kit/kit/log"
	"github.com/goph/emperror"
)

const (
	minMaxWorkers       = 1
	defaultMaxWorkers   = 5
	minMaxBatchSize     = 0
	defaultMaxBatchSize = 1
	minMaxBatchWaitTime = time.Duration(0) * time.Second
	defaultMinQueueSize = 5
)

var (
	defaultLogger = log.NewNopLogger()
)

type BatchInserter struct {
	insertQueue   chan db.Record
	inserter      db.Inserter
	insertWorkers semaphore.Interface
	wg            sync.WaitGroup
	measures      *Measures
	logger        log.Logger
	config        Config
}

type Config struct {
	MaxWorkers       int
	MaxBatchSize     int
	MaxBatchWaitTime time.Duration
	QueueSize        int
}

func NewBatchInserter(config Config, logger log.Logger, metricsRegistry xmetrics.Registry, inserter db.Inserter) (*BatchInserter, error) {
	if inserter == nil {
		return nil, errors.New("invalid inserter")
	}
	if config.MaxWorkers < minMaxWorkers {
		config.MaxWorkers = defaultMaxWorkers
	}
	if config.MaxBatchSize < minMaxBatchSize {
		config.MaxBatchSize = defaultMaxBatchSize
	}
	if config.MaxBatchWaitTime < minMaxBatchWaitTime {
		config.MaxBatchWaitTime = minMaxBatchWaitTime
	}
	if config.QueueSize < defaultMinQueueSize {
		config.QueueSize = defaultMinQueueSize
	}
	if logger == nil {
		logger = defaultLogger
	}

	measures := NewMeasures(metricsRegistry)
	workers := semaphore.New(config.MaxWorkers)
	queue := make(chan db.Record, config.QueueSize)
	b := BatchInserter{
		config:        config,
		logger:        logger,
		measures:      measures,
		insertWorkers: workers,
		inserter:      inserter,
		insertQueue:   queue,
	}
	return &b, nil
}

func (b *BatchInserter) Start() {
	b.wg.Add(1)
	go b.batchRecords()
}

func (b *BatchInserter) Insert(record db.Record) {
	b.insertQueue <- record
	if b.measures != nil {
		b.measures.InsertingQueue.Add(1.0)
	}
}

func (b *BatchInserter) Stop() {
	close(b.insertQueue)
	b.wg.Wait()
}

func (b *BatchInserter) batchRecords() {
	var (
		records      []db.Record
		timeToSubmit time.Time
	)
	defer b.wg.Done()
	for record := range b.insertQueue {
		// if we don't have any records, then this is our first and started
		// the timer until submitting
		if len(records) == 0 {
			timeToSubmit = time.Now().Add(b.config.MaxBatchWaitTime)
		}

		if b.measures != nil {
			b.measures.InsertingQueue.Add(-1.0)
		}
		records = append(records, record)

		// if we have filled up the batch or if we are out of time, we insert
		// what we have
		if (b.config.MaxBatchSize != 0 && len(records) >= b.config.MaxBatchSize) || time.Now().After(timeToSubmit) {
			b.insertWorkers.Acquire()
			go b.insertRecords(records)
			// don't need to remake an array each time, just remove the values
			records = records[:0]
		}

	}

	// Grab all the workers to make sure they are done.
	for i := 0; i < b.config.MaxWorkers; i++ {
		b.insertWorkers.Acquire()
	}
}

func (b *BatchInserter) insertRecords(records []db.Record) {
	defer b.insertWorkers.Release()
	err := b.inserter.InsertRecords(records...)
	if err != nil {
		if b.measures != nil {
			b.measures.DroppedEventsFromDbFailCount.Add(float64(len(records)))
		}
		logging.Error(b.logger, emperror.Context(err)...).Log(logging.MessageKey(),
			"Failed to add records to the database", logging.ErrorKey(), err.Error())
		return
	}
	logging.Debug(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", records)
	logging.Info(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", len(records))
}
