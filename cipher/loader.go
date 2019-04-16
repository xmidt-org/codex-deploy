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

type KeyLoader interface {
	GetBytes() ([]byte, error)
}
type EncryptLoader interface {
	LoadEncrypt() (Encrypt, error)
}
type DecryptLoader interface {
	LoadDecrypt() (Decrypt, error)
}

type FileLoader struct {
	Path string
}

func (f *FileLoader) GetBytes() ([]byte, error) {
	return ioutil.ReadFile(f.Path)
}

type BytesLoader struct {
	Data []byte
}

func (b *BytesLoader) GetBytes() ([]byte, error) {
	return b.Data, nil
}

func GetPrivateKey(loader KeyLoader) (*rsa.PrivateKey, error) {
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

func GetPublicKey(loader KeyLoader) (*rsa.PublicKey, error) {
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

func (o *Options) LoadEncrypt() (Encrypt, error) {
	if o.Logger == nil {
		o.Logger = logging.DefaultLogger()
	}
	logging.Info(o.Logger).Log(logging.MessageKey(), "new encrypter", "options", o)

	if algorithm, ok := o.Algorithm["type"].(string); ok {
		switch strings.ToLower(algorithm) {
		case "noop":
			return DefaultCipherEncrypter(), nil
		case "box":
			if o.SenderPrivateKey == "" || o.RecipientPublicKey == "" {
				break
			}
			boxLoader := BoxLoader{
				PrivateKey: &FileLoader{
					Path: o.SenderPrivateKey,
				},
				PublicKey: &FileLoader{
					Path: o.RecipientPublicKey,
				},
			}
			return boxLoader.LoadEncrypt()
		case "basic":
			if hashName, ok := o.Algorithm["hash"].(string); ok {
				if o.SenderPrivateKey == "" || o.RecipientPublicKey == "" {
					break
				}
				basicLoader := BasicLoader{
					Hash: &BasicHashLoader{hashName},
					PrivateKey: &FileLoader{
						Path: o.SenderPrivateKey,
					},
					PublicKey: &FileLoader{
						Path: o.RecipientPublicKey,
					},
				}
				return basicLoader.LoadEncrypt()
			} else {
				return nil, errors.New("failed to find hash name for basic algorithm cipher type")
			}
		}
	}

	return DefaultCipherEncrypter(), errors.New("failed to load custom algorithm")
}

func (o *Options) LoadDecrypt() (Decrypt, error) {
	if o.Logger == nil {
		o.Logger = logging.DefaultLogger()
	}
	logging.Info(o.Logger).Log(logging.MessageKey(), "new decrypter", "options", o)

	if algorithm, ok := o.Algorithm["type"].(string); ok {
		switch strings.ToLower(algorithm) {
		case "noop":
			return DefaultCipherDecrypter(), nil
		case "box":
			if o.RecipientPrivateKey == "" || o.SenderPublicKey == "" {
				break
			}
			boxLoader := BoxLoader{
				PrivateKey: &FileLoader{
					Path: o.RecipientPrivateKey,
				},
				PublicKey: &FileLoader{
					Path: o.SenderPublicKey,
				},
			}
			return boxLoader.LoadDecrypt()
		case "basic":
			if o.RecipientPrivateKey == "" || o.SenderPublicKey == "" {
				break
			}
			if hashName, ok := o.Algorithm["hash"].(string); ok {
				basicLoader := BasicLoader{
					Hash: &BasicHashLoader{hashName},
					PrivateKey: &FileLoader{
						Path: o.RecipientPrivateKey,
					},
					PublicKey: &FileLoader{
						Path: o.SenderPublicKey,
					},
				}
				return basicLoader.LoadDecrypt()
			} else {
				return nil, errors.New("failed to find hash name for basic algorithm cipher type")
			}
		}
	}

	return DefaultCipherDecrypter(), errors.New("failed to load custom algorithm")
}
