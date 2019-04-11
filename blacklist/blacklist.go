package blacklist

import (
	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"sync"
	"time"
)

type BlackListedItem struct {
	ID     string
	Reason string
}

type List interface {
	InList(ID string) bool
}

type SyncList struct {
	data     map[string]string
	dataLock sync.RWMutex
}

func NewEmptySyncList() SyncList {
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

func (m *SyncList) UpdateList(data []BlackListedItem) {

	newData := make(map[string]string)
	for _, device := range data {
		newData[device.ID] = device.Reason
	}

	m.dataLock.Lock()
	m.data = newData
	m.dataLock.Unlock()
}

type Updater interface {
	GetBlacklist() ([]BlackListedItem, error)
}

type listRefresher struct {
	logger log.Logger

	updater Updater
	cache   SyncList
}

func (d *listRefresher) InList(ID string) bool {
	return d.InList(ID)
}

func (d *listRefresher) updateList() {
	if list, err := d.updater.GetBlacklist(); err != nil {
		d.cache.UpdateList(list)
	} else {
		logging.Error(d.logger).Log(logging.MessageKey(), "failed to update list", logging.ErrorKey(), err)
	}
}

type RefresherConfig struct {
	UpdateInterval time.Duration
	Logger         log.Logger
}

func NewListRefresher(config RefresherConfig, updater Updater, stop chan struct{}) List {
	if config.Logger == nil {
		config.Logger = logging.DefaultLogger()
	}
	listDB := listRefresher{
		logger:  config.Logger,
		updater: updater,
		cache:   NewEmptySyncList(),
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
