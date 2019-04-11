package blacklist

import (
	"github.com/Comcast/codex/db"
	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"sync"
	"time"
)

type List interface {
	InList(ID string) bool
}

type memList struct {
	data     map[string]string
	dataLock sync.RWMutex
}

func newEmptyMemList() memList {
	return memList{
		data: make(map[string]string),
	}
}

func (m *memList) InList(ID string) (val string, ok bool) {
	m.dataLock.RLock()
	val, ok = m.data[ID]
	m.dataLock.RUnlock()
	return
}

func (m *memList) updateList(data []db.BlackDevice) {

	newData := make(map[string]string)
	for _, device := range data {
		newData[device.DeviceID] = device.Reason
	}

	m.dataLock.Lock()
	m.data = newData
	m.dataLock.Unlock()
}

type dbList struct {
	logger log.Logger

	listGetter db.RetryListGService
	cache      memList
}

func (d *dbList) InList(ID string) bool {
	return d.InList(ID)
}

func (d *dbList) updateList() {
	if list, err := d.listGetter.GetBlacklist(); err != nil {
		d.cache.updateList(list)
	} else {
		logging.Error(d.logger).Log(logging.MessageKey(), "failed to update list", logging.ErrorKey(), err)
	}
}

type DBConfig struct {
	UpdateInterval time.Duration
	Logger         log.Logger
}

func NewDBList(config DBConfig, listRetryGetService db.RetryListGService, stop chan struct{}) List {
	if config.Logger == nil {
		config.Logger = logging.DefaultLogger()
	}
	listDB := dbList{
		logger:     config.Logger,
		listGetter: listRetryGetService,
		cache:      newEmptyMemList(),
	}

	go func() {
		ticker := time.NewTicker(config.UpdateInterval)
		for {
			select {
			case <-stop:
				logging.Info(listDB.logger).Log(logging.MessageKey(), "Stopping updater")
				ticker.Stop()
				return
			case <-ticker.C:
				listDB.updateList()
			}
		}
	}()
	logging.Debug(listDB.logger).Log(logging.MessageKey(), "starting db list", "interval", config.UpdateInterval)
	return &listDB
}
