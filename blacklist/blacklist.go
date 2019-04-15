package blacklist

import (
	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"regexp"
	"sync"
	"time"
)

type BlackListedItem struct {
	ID     string
	Reason string
}

func (BlackListedItem) TableName() string {
	return "blacklist"
}

type List interface {
	InList(ID string) (reason string, ok bool)
}

type SyncList struct {
	rules    map[string]string
	dataLock sync.RWMutex
}

func NewEmptySyncList() SyncList {
	return SyncList{
		rules: make(map[string]string),
	}
}

func (m *SyncList) InList(ID string) (string, bool) {
	m.dataLock.RLock()
	defer m.dataLock.RUnlock()

	// fast return of raw string
	if reason, ok := m.rules[ID]; ok {
		return reason, true
	}
	// for regex
	for pattern, reason := range m.rules {
		if matched, err := regexp.MatchString(pattern, ID); err == nil {
			if matched {
				return reason, true
			}
		}
	}
	return "", false
}

func (m *SyncList) UpdateList(data []BlackListedItem) {

	newData := make(map[string]string)
	for _, device := range data {
		newData[device.ID] = device.Reason
	}

	m.dataLock.Lock()
	m.rules = newData
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

func (d *listRefresher) InList(ID string) (string, bool) {
	return d.cache.InList(ID)
}

func (d *listRefresher) updateList() {
	if list, err := d.updater.GetBlacklist(); err == nil {
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
		// do initial update
		listDB.updateList()

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
