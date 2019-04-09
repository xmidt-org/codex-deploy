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

	decrypt, err := LoadPrivate(v)
	assert.NoError(err)
	assert.NotEmpty(decrypt)
}

func TestNOOPViper(t *testing.T) {
	assert := assert.New(t)

	v := viper.New()
	path, err := os.Getwd()
	assert.NoError(err)
	v.AddConfigPath(path)
	v.SetConfigName("noop")

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("%s\n", err)
	}

	encrypt, err := LoadPublic(v)
	assert.NoError(err)

	msg := "hello"
	data, err := encrypt.EncryptMessage([]byte(msg))
	assert.NoError(err)
	assert.Equal([]byte(msg), data)
}
