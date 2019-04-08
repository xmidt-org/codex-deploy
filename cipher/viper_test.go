package cipher

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestViper(t *testing.T) {
	assert := assert.New(t)

	v := viper.New()
	path, err := os.Getwd()
	assert.NoError(err)
	v.AddConfigPath(path)
	v.SetConfigName("example")

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("%s\n", err)
	}

	config, err := Load(v)
	assert.NoError(err)
	assert.NotEmpty(config)
	assert.NotEmpty(config.Key.GetBytes())
}
