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
	"encoding/pem"
	"errors"
)

type BoxLoader struct {
	PrivateKey KeyLoader
	PublicKey  KeyLoader
}

func (boxLoader *BoxLoader) getBoxPrivateKey() ([32]byte, error) {
	var privateKey [32]byte
	data, err := boxLoader.PrivateKey.GetBytes()
	if err != nil {
		return privateKey, nil
	}
	privatePem, _ := pem.Decode(data)
	if privatePem.Type != "BOX PRIVATE KEY" {
		return privateKey, errors.New("incorrect pem type: " + privatePem.Type)
	}
	copy(privateKey[0:32], privatePem.Bytes[:])
	return privateKey, nil
}

func (boxLoader *BoxLoader) getBoxPublicKey() ([32]byte, error) {
	var publicKey [32]byte
	data, err := boxLoader.PublicKey.GetBytes()
	if err != nil {
		return publicKey, nil
	}
	publicPem, _ := pem.Decode(data)
	if publicPem.Type != "BOX PUBLIC KEY" {
		return publicKey, errors.New("incorrect pem type: " + publicPem.Type)
	}
	copy(publicKey[0:32], publicPem.Bytes[:])
	return publicKey, nil
}

func (boxLoader *BoxLoader) LoadEncrypt() (Encrypt, error) {
	publicKey, err := boxLoader.getBoxPublicKey()
	if err != nil {
		return nil, err
	}

	privateKey, err := boxLoader.getBoxPrivateKey()
	if err != nil {
		return nil, err
	}
	return NewBoxEncrypter(privateKey, publicKey), nil
}

func (boxLoader *BoxLoader) LoadDecrypt() (Decrypt, error) {
	publicKey, err := boxLoader.getBoxPublicKey()
	if err != nil {
		return nil, err
	}

	privateKey, err := boxLoader.getBoxPrivateKey()
	if err != nil {
		return nil, err
	}
	return NewBoxDecrypter(privateKey, publicKey), nil
}
