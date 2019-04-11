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

type SyncList struct {
	data     map[string]string
	dataLock sync.RWMutex
}

func NewEmptyMemList() SyncList {
	return SyncList{
		data: make(map[string]string),
	}
}

func (m *SyncList) InList(ID string) (val string, ok bool) {
	m.dataLock.RLock()
	val, ok = m.data[ID]
	m.dataLock.RUnlock()
	return
}

func (m *SyncList) UpdateList(data []db.BlacklistedDevice) {

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

	listGetter db.ListGetter
	cache      SyncList
}

func (d *dbList) InList(ID string) bool {
	return d.InList(ID)
}

func (d *dbList) updateList() {
	if list, err := d.listGetter.GetBlacklist(); err != nil {
		d.cache.UpdateList(list)
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
		cache:      NewEmptyMemList(),
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
