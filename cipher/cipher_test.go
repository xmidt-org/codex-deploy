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

	publicCrypter := NewPublicCrypter(crypto.BLAKE2b_512, &key.PublicKey)
	assert.NotEmpty(publicCrypter)
	privateCrypter := NewPrivateCrypter(crypto.BLAKE2b_512, key)
	assert.NotEmpty(privateCrypter)

	message := []byte("Hello World")

	encodedMSG, err := publicCrypter.EncryptMessage(message)
	assert.NoError(err)
	assert.NotEmpty(encodedMSG)

	msg, err := privateCrypter.DecryptMessage(encodedMSG)
	assert.NoError(err)
	assert.Equal(message, msg)
}

func TestLargeKey(t *testing.T) {
	assert := assert.New(t)

	key := GeneratePrivateKey(^int(0))
	assert.NotNil(key)

	publicCrypter := NewPublicCrypter(crypto.SHA1, &key.PublicKey)
	assert.NotEmpty(publicCrypter)

	message := []byte("Hello World")

	_, err := publicCrypter.EncryptMessage(message)
	assert.Error(err)
}

func TestSmallKey(t *testing.T) {
	assert := assert.New(t)

	key := GeneratePrivateKey(64)
	assert.NotNil(key)

	publicCrypter := NewPublicCrypter(crypto.SHA512, &key.PublicKey)
	assert.NotEmpty(publicCrypter)

	message := []byte("Hello World")

	_, err := publicCrypter.EncryptMessage(message)
	assert.Error(err)
}

type testData struct {
	message     string
	expectedErr bool
}

func TestCrypters(t *testing.T) {
	keyA := GeneratePrivateKey(2048)
	sha512 := NewPublicCrypter(crypto.SHA512, &keyA.PublicKey)
	sha512Private := NewPrivateCrypter(crypto.SHA512, keyA)

	keyB := GeneratePrivateKey(2048)
	blake512 := NewPublicCrypter(crypto.BLAKE2b_512, &keyB.PublicKey)
	blake512Private := NewPrivateCrypter(crypto.BLAKE2b_512, keyB)

	keyC := GeneratePrivateKey(4096)
	largeMD5 := NewPublicCrypter(crypto.BLAKE2b_512, &keyC.PublicKey)
	largeMD5Private := NewPrivateCrypter(crypto.BLAKE2b_512, keyC)

	crpyters := []struct {
		publicCrypter  PublicKeyCipher
		privateCrypter PrivateKeyCipher
		name           string
	}{
		{new(NOOP), new(NOOP), "noop"},
		{sha512, sha512Private, "sha512"},
		{blake512, blake512Private, "blake512"},
		{largeMD5, largeMD5Private, "largeMD5"},
	}

	testData := []testData{
		{"Hello World", false},
		{"Hello, 世界", false},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", true},
	}

	for _, c := range crpyters {
		t.Run(c.name, func(t *testing.T) {
			testCrypterOrder(t, c.publicCrypter, c.privateCrypter, testData)
		})
	}
}

func testCrypterOrder(t *testing.T, publicCipher PublicKeyCipher, privateCipher PrivateKeyCipher, data []testData) {
	assert := assert.New(t)

	encodedMSGS := make([][]byte, len(data))
	signatures := make([][]byte, len(data))

	for index, item := range data {
		// Encode Message
		encodedMSG, err := publicCipher.EncryptMessage([]byte(item.message))
		if err != nil && item.expectedErr {
			assert.Empty(encodedMSG)
			assert.Contains(err.Error(), "too long for RSA public key size")
			continue
		}
		assert.NoError(err)
		assert.NotEmpty(encodedMSG)
		encodedMSGS[index] = encodedMSG

		// Sign Message
		signature, err := privateCipher.Sign([]byte(item.message))
		assert.NoError(err)
		assert.NotEmpty(signature)
		signatures[index] = signature

	}

	for index, encodedMSG := range encodedMSGS {
		if len(encodedMSG) == 0 {
			continue
		}
		// Decode Message
		msg, err := privateCipher.DecryptMessage(encodedMSG)
		assert.NoError(err)
		assert.Equal([]byte(data[index].message), msg)

		// Verify Message
		verified := publicCipher.VerifyMessage(msg, signatures[index])
		assert.True(verified)
	}
}
