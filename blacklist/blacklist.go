package blacklist

import (
	"regexp"
	"sync"
	"time"

	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
)

const (
	defaultUpdateInterval = time.Minute
)

// BlackListedItem is the regex that expresses the devices that are blacklisted
// and the reason why.
type BlackListedItem struct {
	ID     string
	Reason string
}

// TableName sets BlackListedItem's table name to be "blacklist"; for the GORM driver.
func (BlackListedItem) TableName() string {
	return "blacklist"
}

// List is for checking if a device id is in the blacklist.
type List interface {
	InList(ID string) (reason string, ok bool)
}

// SyncList is an implemention of the List interface that works synchronously.
type SyncList struct {
	rules    map[string]string
	dataLock sync.RWMutex
}

// NewEmptySyncList creates a new SyncList that holds no information.
func NewEmptySyncList() SyncList {
	return SyncList{
		rules: make(map[string]string),
	}
}

// InList returns whether or not a device is on the blacklist and why, if it's
// on the list.
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

// UpdateList takes the data given and overwrites the blacklist with the new
// information.
func (m *SyncList) UpdateList(data []BlackListedItem) {

	newData := make(map[string]string)
	for _, device := range data {
		newData[device.ID] = device.Reason
	}

	m.dataLock.Lock()
	m.rules = newData
	m.dataLock.Unlock()
}

// Updater is for getting the blacklist.
type Updater interface {
	GetBlacklist() ([]BlackListedItem, error)
}

type listRefresher struct {
	logger log.Logger

	updater Updater
	cache   SyncList
}

// InList checks if a specified device id is on the blacklist.
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

// RefresherConfig is the configuration specifying how often to update the list
// and what logger to use when logging.
type RefresherConfig struct {
	UpdateInterval time.Duration
	Logger         log.Logger
}

// NewListRefresher takes the given values and uses them to create a new listRefresher
func NewListRefresher(config RefresherConfig, updater Updater, stop chan struct{}) List {
	if config.Logger == nil {
		config.Logger = logging.DefaultLogger()
	}
	if config.UpdateInterval == time.Duration(0)*time.Second {
		config.UpdateInterval = defaultUpdateInterval
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
