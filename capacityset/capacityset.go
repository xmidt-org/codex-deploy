/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */


package capacityset

import (
	"github.com/Comcast/webpa-common/semaphore"
	"sync"
)

// Set is the interface of the common set storage pattern
type Set interface {
	// Adds an element to the set. Returns whether
	// the item was added.
	Add(item interface{}) bool

	// Pop returns an item from the set
	Pop() interface{}

	// Size returns the number of items in the set
	Size() int
}

// NewCapacitySet returns the set interface with max capacity
// allowing of a set with cap
func NewCapacitySet(capacity int) Set {
	return &limitedSet{
		limit: semaphore.New(capacity),
		data:  map[interface{}]bool{},
	}
}

type limitedSet struct {
	limit semaphore.Interface
	sync.RWMutex
	data map[interface{}]bool
}

func (set *limitedSet) Add(item interface{}) bool {
	set.limit.Acquire()
	set.Lock()
	defer set.Unlock()

	if set.data[item] {
		return false //False if it existed already
	}
	set.data[item] = true
	return true
}

func (set *limitedSet) Pop() interface{} {
	set.Lock()
	defer func() {
		set.Unlock()
		set.limit.Release()
	}()

	for item := range set.data {
		delete(set.data, item)
		return item
	}
	return nil
}

func (set *limitedSet) Size() int {
	set.RLock()
	defer set.RUnlock()

	return len(set.data)
}
