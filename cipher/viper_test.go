package cipher

import (
	"fmt"
	"github.com/Comcast/webpa-common/logging"
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

	encrypter, err := options.GetEncrypter(logging.NewTestLogger(nil, t))
	assert.NoError(err)
	assert.NotNil(encrypter)

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

	encrypter, err := options.GetEncrypter(logging.NewTestLogger(nil, t))

	msg := "hello"
	data, _, err := encrypter.EncryptMessage([]byte(msg))
	assert.NoError(err)
	assert.Equal([]byte(msg), data)
}

func TestBoxBothSides(t *testing.T) {
	assert := assert.New(t)

	vSend := viper.New()
	path, err := os.Getwd()
	assert.NoError(err)
	vSend.AddConfigPath(path)
	vSend.SetConfigName("boxSender")
	if err := vSend.ReadInConfig(); err != nil {
		t.Fatalf("%s\n", err)
	}

	options, err := FromViper(vSend)
	assert.NoError(err)

	encrypter, err := options.GetEncrypter(logging.NewTestLogger(nil, t))
	assert.NoError(err)

	vRec := viper.New()
	assert.NoError(err)
	vRec.AddConfigPath(path)
	vRec.SetConfigName("boxRecipient")
	if err := vRec.ReadInConfig(); err != nil {
		t.Fatalf("%s\n", err)
	}

	options, err = FromViper(vRec)
	assert.NoError(err)

	decrypters := PopulateCiphers(options, logging.NewTestLogger(nil, t))

	assert.NoError(err)

	msg := []byte("hello")
	data, nonce, err := encrypter.EncryptMessage(msg)
	assert.NoError(err)

	if decrypter, ok := decrypters.Get(encrypter.GetAlgorithm(), encrypter.GetKID()); ok {
		decodedMSG, err := decrypter.DecryptMessage(data, nonce)
		assert.NoError(err)

		assert.Equal(msg, decodedMSG)
	} else {
		assert.Fail("failed to get decrypter with kid")
	}
}

func TestGetDecrypterErr(t *testing.T) {
	assert := assert.New(t)

	vSend := viper.New()
	path, err := os.Getwd()
	assert.NoError(err)
	vSend.AddConfigPath(path)
	vSend.SetConfigName("boxRecipient")
	if err := vSend.ReadInConfig(); err != nil {
		t.Fatalf("%s\n", err)
	}

	options, err := FromViper(vSend)
	assert.NoError(err)

	decrypters := PopulateCiphers(options, logging.NewTestLogger(nil, t))
	fmt.Printf("%#v\n", decrypters)

	decrypter, ok := decrypters.Get(Box, "test")
	assert.True(ok)
	assert.NotNil(decrypter)

	decrypter, ok = decrypters.Get(None, "none")
	assert.True(ok)
	assert.NotNil(decrypter)

	// negative test
	decrypter, ok = decrypters.Get(None, "neato")
	assert.False(ok)
	assert.Nil(decrypter)

	decrypter, ok = decrypters.Get(RSAAsymmetric, "testing")
	assert.False(ok)
	assert.Nil(decrypter)
}
