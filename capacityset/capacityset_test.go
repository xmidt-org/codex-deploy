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
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestDuplicateInsert(t *testing.T) {
	assert := assert.New(t)

	set := NewCapacitySet(5)
	assert.True(set.Add(5))
	assert.False(set.Add(5))
}

func TestPop(t *testing.T) {

	assert := assert.New(t)

	set := NewCapacitySet(5)
	assert.True(set.Add(5))
	assert.True(set.Add(1))
	assert.True(set.Add(3))
	assert.True(set.Add(2))

	assert.Equal(4, set.Size())
	assert.NotNil(set.Pop())
}

func TestBlockingAdd(t *testing.T) {
	assert := assert.New(t)

	set := NewCapacitySet(5)
	assert.True(set.Add(1))
	assert.True(set.Add(2))
	assert.True(set.Add(3))
	assert.True(set.Add(4))
	assert.True(set.Add(5))

	wg := new(sync.WaitGroup)
	wg.Add(1)

	called := false
	go func() {
		set.Add(10)
		called = true
		wg.Done()

	}()
	assert.NotNil(set.Pop())
	wg.Wait()
	assert.True(called)
}
