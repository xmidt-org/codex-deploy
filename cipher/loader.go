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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

var (
	hashFunctions = map[string]crypto.Hash{
		"BLAKE2B512": crypto.BLAKE2b_512,
		"SHA1":       crypto.SHA1,
		"SHA512":     crypto.SHA512,
		"MD5":        crypto.MD5,
	}
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

type KeyLoader interface {
	GetBytes() ([]byte, error)
}

type FileLoader struct {
	Path string
}

func (f *FileLoader) GetBytes() ([]byte, error) {
	return ioutil.ReadFile(f.Path)
}

type BasicLoader struct {
	Hash       HashLoader
	PrivateKey KeyLoader
	PublicKey  KeyLoader
}

type BoxLoader struct {
	PrivateKey KeyLoader
	PublicKey  KeyLoader
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

func getBoxPrivateKey(loader KeyLoader) ([32]byte, error) {
	var privateKey [32]byte
	data, err := loader.GetBytes()
	if err != nil {
		return privateKey, nil
	}
	privatePem, _ := pem.Decode(data)
	if privatePem.Type != "BOX PRIVATE KEY" {
		return privateKey, errors.New("incorrect pem type: " + privatePem.Type)
	}
	copy(privateKey[0:24], privatePem.Bytes[:])
	return privateKey, nil
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
		return nil, emperror.Wrap(err, "failed to load public key x509.ParsePKCS1PublicKey")
	}

	if publicKey, ok := parsedKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("failed convert parsed key to public key")
	} else {
		return publicKey, nil
	}
}

func getBoxPublicKey(loader KeyLoader) ([32]byte, error) {
	var publicKey [32]byte
	data, err := loader.GetBytes()
	if err != nil {
		return publicKey, nil
	}
	publicPem, _ := pem.Decode(data)
	if publicPem.Type != "BOX PUBLIC KEY" {
		return publicKey, errors.New("incorrect pem type: " + publicPem.Type)
	}
	copy(publicKey[0:24], publicPem.Bytes[:])
	return publicKey, nil
}

func LoadBasicEncrypter(loader BasicLoader) (Encrypt, error) {
	hashFunc, err := loader.Hash.GetHash()
	if err != nil {
		return nil, err
	}

	publicKey, err := getPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := getPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBasicEncrypter(hashFunc, privateKey, publicKey), nil
}

func LoadBoxEncrypter(loader BoxLoader) (Encrypt, error) {
	publicKey, err := getBoxPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := getBoxPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBoxEncrypter(privateKey, publicKey), nil
}

func LoadBasicDecrypter(loader BasicLoader) (Decrypt, error) {
	hashFunc, err := loader.Hash.GetHash()
	if err != nil {
		return nil, err
	}

	publicKey, err := getPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := getPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBasicDecrypter(hashFunc, privateKey, publicKey), nil
}

func LoadBoxDecrypter(loader BoxLoader) (Decrypt, error) {
	publicKey, err := getBoxPublicKey(loader.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := getBoxPrivateKey(loader.PrivateKey)
	if err != nil {
		return nil, err
	}
	return NewBoxDecrypter(privateKey, publicKey), nil
}

type Options struct {
	// Logger is the go-kit Logger to use for server startup and error logging.  If not
	// supplied, logging.DefaultLogger() is used instead.
	Logger log.Logger `json:"-"`

	Algorithm map[string]interface{} `json:"algorithm,omitempty"`

	SenderPrivateKey    string `json:"senderPrivateKey,omitempty"`
	SenderPublicKey     string `json:"senderPublicKey,omitempty"`
	RecipientPrivateKey string `json:"recipientPrivateKey,omitempty"`
	RecipientPublicKey  string `json:"recipientPublicKey,omitempty"`
}

func NewEncrypter(o Options) (Encrypt, error) {
	if o.Logger == nil {
		o.Logger = logging.DefaultLogger()
	}

	logging.Info(o.Logger).Log(logging.MessageKey(), "new encrypter", "options", o)
	if algorithm, ok := o.Algorithm["type"].(string); ok {
		switch strings.ToLower(algorithm) {
		case "noop":
			return DefaultCipherEncrypter(), nil
		case "box":
			return LoadBoxEncrypter(BoxLoader{
				PrivateKey: &FileLoader{
					Path: o.SenderPrivateKey,
				},
				PublicKey: &FileLoader{
					Path: o.RecipientPublicKey,
				},
			})
		case "basic":
			if hashName, ok := o.Algorithm["hash"].(string); ok {
				return LoadBasicEncrypter(BasicLoader{
					Hash: &BasicHashLoader{hashName},
					PrivateKey: &FileLoader{
						Path: o.SenderPrivateKey,
					},
					PublicKey: &FileLoader{
						Path: o.RecipientPublicKey,
					},
				})
			} else {
				return nil, errors.New("failed to find hash name for basic algorithm cipher type")
			}
		}
	}

	return DefaultCipherEncrypter(), errors.New("failed to load custom algorithm")
}

func NewDecrypter(o Options) (Decrypt, error) {
	if o.Logger == nil {
		o.Logger = logging.DefaultLogger()
	}
	logging.Info(o.Logger).Log(logging.MessageKey(), "new decrypter", "options", o)

	if algorithm, ok := o.Algorithm["type"].(string); ok {
		switch strings.ToLower(algorithm) {
		case "noop":
			return DefaultCipherDecrypter(), nil
		case "box":
			return LoadBoxDecrypter(BoxLoader{
				PrivateKey: &FileLoader{
					Path: o.RecipientPrivateKey,
				},
				PublicKey: &FileLoader{
					Path: o.SenderPublicKey,
				},
			})
		case "basic":
			if hashName, ok := o.Algorithm["hash"].(string); ok {
				return LoadBasicDecrypter(BasicLoader{
					Hash: &BasicHashLoader{hashName},
					PrivateKey: &FileLoader{
						Path: o.RecipientPrivateKey,
					},
					PublicKey: &FileLoader{
						Path: o.SenderPublicKey,
					},
				})
			} else {
				return nil, errors.New("failed to find hash name for basic algorithm cipher type")
			}
		}
	}

	return DefaultCipherDecrypter(), errors.New("failed to load custom algorithm")
}
