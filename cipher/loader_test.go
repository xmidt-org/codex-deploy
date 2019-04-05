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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func TestLoadPublicKey(t *testing.T) {

}
func TestCipherLoader(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.Getwd()
	assert.NoError(err)

	privateCrypter, err := LoadPrivateKey(Config{
		Hash: "SHA512",
		Key: &FileLoader{
			Path: dir + string(os.PathSeparator) +"private.pem",
		},
	})
	assert.NotEmpty(privateCrypter)
	assert.NoError(err)


	publicCrypter, err := LoadPublicKey(Config{
		Hash: "SHA512",
		Key: &FileLoader{
			Path: dir + string(os.PathSeparator) + "public.pem",
		},
	})
	assert.NotEmpty(publicCrypter)
	assert.NoError(err)

	message := []byte("Hello World")

	encodedMSG, err := publicCrypter.EncryptMessage(message)
	assert.NoError(err)
	assert.NotEmpty(encodedMSG)

	msg, err := privateCrypter.DecryptMessage(encodedMSG)
	assert.NoError(err)
	assert.Equal(message, msg)
}
