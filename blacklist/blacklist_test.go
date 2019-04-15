package blacklist

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlacklist(t *testing.T) {
	assert := assert.New(t)
	list := NewEmptySyncList()

	bad := "all bad"
	singleBad := "Bad Ideas"

	list.UpdateList([]BlackListedItem{{"randomID", singleBad}, {"bad.*", bad}})

	reason, ok := list.InList("randomID")
	assert.True(ok)
	assert.Equal(singleBad, reason)

	reason, ok = list.InList("badDevice")
	assert.True(ok)
	assert.Equal(bad, reason)

	reason, ok = list.InList("badIdea")
	assert.True(ok)
	assert.Equal(bad, reason)

	reason, ok = list.InList("happyDevice")
	assert.False(ok)
	assert.Empty(reason)
}
