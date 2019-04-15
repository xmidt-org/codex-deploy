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

	options, err := FromViper(v)
	assert.NoError(err)

	encrypter, err := NewEncrypter(*options)
	assert.NoError(err)
	assert.NotEmpty(options)

	msg := "hello"
	data, _, err := encrypter.EncryptMessage([]byte(msg))
	assert.NoError(err)
	assert.NotEqual([]byte(msg), data)
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

	options, err := FromViper(v)
	assert.NoError(err)

	encrypter, err := NewEncrypter(*options)

	msg := "hello"
	data, _, err := encrypter.EncryptMessage([]byte(msg))
	assert.NoError(err)
	assert.Equal([]byte(msg), data)
}
