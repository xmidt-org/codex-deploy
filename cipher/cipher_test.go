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

package cipher

import (
	"crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicEncrypt(t *testing.T) {
	assert := assert.New(t)

	key := GeneratePrivateKey(2048)
	assert.NotNil(key)

	crypter := NewCrypter(crypto.BLAKE2b_512, key)
	assert.NotEmpty(crypter)

	message := "Hello World"

	encodedMSG, err := crypter.EncryptMessage(message)
	assert.NoError(err)
	assert.NotEmpty(encodedMSG)

	msg, err := crypter.DecryptMessage(encodedMSG)
	assert.NoError(err)
	assert.Equal(message, msg)
}

func TestLargeKey(t *testing.T) {
	assert := assert.New(t)

	key := GeneratePrivateKey(^int(0))
	assert.NotNil(key)

	crypter := NewCrypter(crypto.SHA1, key)
	assert.NotEmpty(crypter)

	message := "Hello World"

	_, err := crypter.EncryptMessage(message)
	assert.Error(err)
}

func TestSmallKey(t *testing.T) {
	assert := assert.New(t)

	key := GeneratePrivateKey(64)
	assert.NotNil(key)

	crypter := NewCrypter(crypto.SHA512, key)
	assert.NotEmpty(crypter)

	message := "Hello World"

	_, err := crypter.EncryptMessage(message)
	assert.Error(err)
}

type testData struct {
	message     string
	expectedErr bool
}

func TestCrypters(t *testing.T) {
	sha512 := NewCrypter(crypto.SHA512, GeneratePrivateKey(2048))
	blake512 := NewCrypter(crypto.BLAKE2b_512, GeneratePrivateKey(2048))
	largeMD5 := NewCrypter(crypto.MD5, GeneratePrivateKey(4096))

	crpyters := []struct {
		crypter Interface
		name    string
	}{
		{new(NOOP), "noop"},
		{sha512, "sha512"},
		{blake512, "blake512"},
		{largeMD5, "largeMD5"},
	}

	testData := []testData{
		{"Hello World", false},
		{"Hello, 世界", false},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", true},
	}

	for _, c := range crpyters {
		t.Run(c.name, func(t *testing.T) {
			testCrypterOrder(t, c.crypter, testData)
		})
	}
}

func testCrypterOrder(t *testing.T, crypter Interface, data []testData) {
	assert := assert.New(t)

	encodedMSGS := make([][]byte, len(data))
	signatures := make([][]byte, len(data))

	for index, item := range data {
		// Encode Message
		encodedMSG, err := crypter.EncryptMessage(item.message)
		if err != nil && item.expectedErr {
			assert.Empty(encodedMSG)
			assert.Contains(err.Error(), "too long for RSA public key size")
			continue
		}
		assert.NoError(err)
		assert.NotEmpty(encodedMSG)
		encodedMSGS[index] = encodedMSG

		// Sign Message
		signature, err := crypter.Sign(item.message)
		assert.NoError(err)
		assert.NotEmpty(signature)
		signatures[index] = signature

	}

	for index, encodedMSG := range encodedMSGS {
		if len(encodedMSG) == 0 {
			continue
		}
		// Decode Message
		msg, err := crypter.DecryptMessage(encodedMSG)
		assert.NoError(err)
		assert.Equal(data[index].message, msg)

		// Verify Message
		verified := crypter.VerifyMessage(msg, signatures[index])
		assert.True(verified)
	}
}
