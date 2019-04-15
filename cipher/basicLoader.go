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
	"errors"
	"strings"
)

func GetHash(hashType string) crypto.Hash {
	if elem, ok := hashFunctions[strings.ToUpper(hashType)]; ok {
		return elem
	}
	return crypto.BLAKE2b_512
}

type HashLoader interface {
	GetHash() (crypto.Hash, error)
}
type BasicHashLoader struct {
	HashName string `mapstructure:"hash"`
}

// GetHash return the given hash from hashFunctions if not found it will return an error.
//   0 is an invalid hash
func (b *BasicHashLoader) GetHash() (crypto.Hash, error) {
	if elem, ok := hashFunctions[strings.ToUpper(b.HashName)]; ok {
		if elem.Available() {
			return elem, nil
		}
		return 0, errors.New("hash " + b.HashName + " is not linked in binary")
	}
	return 0, errors.New("hashname " + b.HashName + " not found")
}

type BasicLoader struct {
	Hash       HashLoader
	PrivateKey KeyLoader
	PublicKey  KeyLoader
}

func (loader *BasicLoader) LoadEncrypt() (Encrypt, error) {
	hashFunc, err := loader.Hash.GetHash()
	if err != nil {
		return nil, err
	}

	publicKey, err := GetPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := GetPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBasicEncrypter(hashFunc, privateKey, publicKey), nil
}

func (loader *BasicLoader) LoadDecrypt() (Decrypt, error) {
	hashFunc, err := loader.Hash.GetHash()
	if err != nil {
		return nil, err
	}

	publicKey, err := GetPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := GetPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBasicDecrypter(hashFunc, privateKey, publicKey), nil
}
