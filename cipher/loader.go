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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"io/ioutil"
)

type KeyLoader interface {
	GetBytes() ([]byte, error)
}

type FileLoader struct {
	Path string
}

func (f *FileLoader) GetBytes() ([]byte, error) {
	return ioutil.ReadFile(f.Path)
}

type Config struct {
	Hash string
	Key  KeyLoader
}

func getPrivateKey(loader KeyLoader) (*rsa.PrivateKey, error) {
	data, err := loader.GetBytes()
	if err != nil {
		return nil, err
	}
	privPem, _ := pem.Decode(data)
	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("incorrect pem type: " + privPem.Type)
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		return nil, err
	}

	if privateKey, ok := parsedKey.(*rsa.PrivateKey); !ok {
		return nil, errors.New("failed convert parsed key to private key")
	} else {
		return privateKey, nil
	}
}

func getPublicKey(loader KeyLoader) (*rsa.PublicKey, error) {
	data, err := loader.GetBytes()
	if err != nil {
		return nil, err
	}
	publicPem, _ := pem.Decode(data)
	if publicPem.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("incorrect pem type: " + publicPem.Type)
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PublicKey(publicPem.Bytes); err != nil {
		return nil, errors.New("x509.ParsePKCS1PublicKey error: " + err.Error())
	}

	if publicKey, ok := parsedKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("failed convert parsed key to public key")
	} else {
		return publicKey, nil
	}
}

func LoadPublicKey(config Config) (PublicKeyCipher, error) {
	hashType := GetHash(config.Hash)
	if hashType == Unknown {
		return nil, errors.New("unknown hash function: " + config.Hash)
	}
	key, err := getPublicKey(config.Key)
	if err != nil {
		return nil, err
	}
	return NewPublicCrypter(hashType.GetHash(), key), nil
}

func LoadPrivateKey(config Config) (PrivateKeyCipher, error) {
	hashType := GetHash(config.Hash)
	if hashType == Unknown {
		return nil, errors.New("unknown hash function: " + config.Hash)
	}
	key, err := getPrivateKey(config.Key)
	if err != nil {
		return nil, err
	}
	return NewPrivateCrypter(hashType.GetHash(), key), nil
}
