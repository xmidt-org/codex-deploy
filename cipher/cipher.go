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

// cipher package is a helper package for encrypting and decrypting messages
package cipher

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
	"hash"
	"io"
	"os"
)

func init() {
	// register crypto.BLAKE2b_512 hash
	crypto.RegisterHash(crypto.BLAKE2b_512, func() hash.Hash {
		b2b, err := blake2b.New512([]byte("73 is the best number"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error from blake2b.New512: %s\n", err)
			return nil
		}
		return b2b
	})
}

type Identification interface {
	// GetAlgorithm will return the algorithm Encrypt and Decrypt uses
	GetAlgorithm() AlgorithmType

	// GetKID returns the id of the specific keys used
	GetKID() string
}

// Encrypt represents the ability to encrypt messages
type Encrypt interface {
	Identification

	// EncryptMessage attempts to encode the message into an array of bytes.
	// and error will be returned if failed to encode the message.
	EncryptMessage(message []byte) (crypt []byte, nonce []byte, err error)
}

// Decrypt represents the ability to decrypt messages
type Decrypt interface {
	Identification

	// DecryptMessage attempts to decode the message into a string.
	// and error will be returned if failed to decode the message.
	DecryptMessage(cipher []byte, nonce []byte) (message []byte, err error)
}

// GeneratePrivateKey will create a private key with the size given
// size must be greater than 64 or else it will default to 64.
//
// Careful with the size, if its too large it won't encrypt the message or take forever
func GeneratePrivateKey(size int) *rsa.PrivateKey {
	if size < 64 {
		// size is to small and it will be hard to find prime numbers
		size = 64
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, size)
	return privateKey
}

func DefaultCipherEncrypter() Encrypt {
	return &NOOP{}
}

func DefaultCipherDecrypter() Decrypt {
	return &NOOP{}
}

// NOOP will just return the message
type NOOP struct{}

func (*NOOP) GetAlgorithm() AlgorithmType {
	return None
}

func (*NOOP) GetKID() string {
	return "none"
}

func (*NOOP) EncryptMessage(message []byte) (crypt []byte, nonce []byte, err error) {
	return message, []byte{}, nil
}

func (*NOOP) DecryptMessage(cipher []byte, nonce []byte) (message []byte, err error) {
	return cipher, nil
}

type rsaEncrypter struct {
	kid                string
	hasher             crypto.Hash
	senderPrivateKey   *rsa.PrivateKey
	recipientPublicKey *rsa.PublicKey
	label              []byte
}

func (c *rsaEncrypter) GetAlgorithm() AlgorithmType {
	if c.recipientPublicKey == nil {
		return RSASymmetric
	}
	return RSAAsymmetric
}

func (c *rsaEncrypter) GetKID() string {
	return c.kid
}

type rsaDecrypter struct {
	kid                 string
	hasher              crypto.Hash
	recipientPrivateKey *rsa.PrivateKey
	senderPublicKey     *rsa.PublicKey
	label               []byte
}

func (c *rsaDecrypter) GetAlgorithm() AlgorithmType {
	if c.senderPublicKey == nil {
		return RSASymmetric
	}
	return RSAAsymmetric
}

func (c *rsaDecrypter) GetKID() string {
	return c.kid
}

func NewRSAEncrypter(hash crypto.Hash, senderPrivateKey *rsa.PrivateKey, recipientPublicKey *rsa.PublicKey, kid string) Encrypt {
	return &rsaEncrypter{
		kid:                kid,
		hasher:             hash,
		senderPrivateKey:   senderPrivateKey,
		recipientPublicKey: recipientPublicKey,
		label:              []byte("codex-rsa-cipher"),
	}
}

func NewRSADecrypter(hash crypto.Hash, recipientPrivateKey *rsa.PrivateKey, senderPublicKey *rsa.PublicKey, kid string) Decrypt {
	return &rsaDecrypter{
		kid:                 kid,
		hasher:              hash,
		recipientPrivateKey: recipientPrivateKey,
		senderPublicKey:     senderPublicKey,
		label:               []byte("codex-rsa-cipher"),
	}
}

func (c *rsaEncrypter) EncryptMessage(message []byte) ([]byte, []byte, error) {
	cipherdata, err := rsa.EncryptOAEP(
		c.hasher.New(),
		rand.Reader,
		c.recipientPublicKey,
		message,
		c.label,
	)
	if err != nil {
		return []byte(""), []byte{}, emperror.Wrap(err, "failed to encrypt message")
	}

	signature := []byte{}

	if c.senderPrivateKey != nil {
		var opts rsa.PSSOptions
		opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

		pssh := c.hasher.New()
		pssh.Write(message)
		hashed := pssh.Sum(nil)

		signature, err = rsa.SignPSS(rand.Reader, c.senderPrivateKey, c.hasher, hashed, &opts)
		if err != nil {
			return []byte(""), []byte{}, emperror.Wrap(err, "failed to sign message")
		}
	}

	return cipherdata, signature, nil
}

func (c *rsaDecrypter) DecryptMessage(cipher []byte, nonce []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(
		c.hasher.New(),
		rand.Reader,
		c.recipientPrivateKey,
		cipher,
		c.label,
	)
	if err != nil {
		return []byte{}, emperror.Wrap(err, "failed to decrypt message")
	}

	if c.senderPublicKey != nil {
		var opts rsa.PSSOptions
		opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

		pssh := c.hasher.New()
		pssh.Write(decrypted)
		hashed := pssh.Sum(nil)

		err = rsa.VerifyPSS(c.senderPublicKey, c.hasher, hashed, nonce, &opts)
		if err != nil {
			return []byte{}, emperror.Wrap(err, "failed to validate signature")
		}
	}

	return decrypted, nil
}

type encryptBox struct {
	kid                string
	senderPrivateKey   [32]byte
	recipientPublicKey [32]byte
	sharedEncryptKey   *[32]byte
}

func (enBox *encryptBox) GetAlgorithm() AlgorithmType {
	return Box
}

func (enBox *encryptBox) GetKID() string {
	return enBox.kid
}

func NewBoxEncrypter(senderPrivateKey [32]byte, recipientPublicKey [32]byte, kid string) Encrypt {

	encrypter := encryptBox{
		kid:                kid,
		senderPrivateKey:   senderPrivateKey,
		recipientPublicKey: recipientPublicKey,
		sharedEncryptKey:   new([32]byte),
	}

	box.Precompute(encrypter.sharedEncryptKey, &encrypter.recipientPublicKey, &encrypter.senderPrivateKey)

	return &encrypter
}

func (enBox *encryptBox) EncryptMessage(message []byte) ([]byte, []byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return []byte(""), []byte{}, emperror.Wrap(err, "failed to generate nonce")
	}

	encrypted := box.SealAfterPrecomputation(nil, message, &nonce, enBox.sharedEncryptKey)

	return encrypted, nonce[:], nil
}

type decryptBox struct {
	kid                 string
	recipientPrivateKey [32]byte
	senderPublicKey     [32]byte
	sharedDecryptKey    *[32]byte
}

func (deBox *decryptBox) GetAlgorithm() AlgorithmType {
	return Box
}

func (deBox *decryptBox) GetKID() string {
	return deBox.kid
}

func NewBoxDecrypter(recipientPrivateKey [32]byte, senderPublicKey [32]byte, kid string) Decrypt {

	decrypter := decryptBox{
		kid:                 kid,
		recipientPrivateKey: recipientPrivateKey,
		senderPublicKey:     senderPublicKey,
		sharedDecryptKey:    new([32]byte),
	}

	box.Precompute(decrypter.sharedDecryptKey, &decrypter.senderPublicKey, &decrypter.recipientPrivateKey)

	return &decrypter
}

func (deBox *decryptBox) DecryptMessage(cipher []byte, nonce []byte) ([]byte, error) {
	var decryptNonce [24]byte
	copy(decryptNonce[:], nonce[:24])

	decrypted, ok := box.OpenAfterPrecomputation(nil, cipher, &decryptNonce, deBox.sharedDecryptKey)
	if !ok {
		return []byte(""), errors.New("failed to decrypt message")
	}

	return decrypted, nil
}
