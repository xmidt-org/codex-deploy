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
	crypto_rand "crypto/rand" // Custom so it's clear which rand we're using.
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

// Encrypt represents the ability to encrypt messages
type Encrypt interface {
	// EncryptMessage attempts to encode the message into an array of bytes.
	// and error will be returned if failed to encode the message.
	EncryptMessage(message []byte) (crypt []byte, nonce []byte, err error)
}

// Decrypt represents the ability to decrypt messages
type Decrypt interface {
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

	privateKey, _ := rsa.GenerateKey(crypto_rand.Reader, size)
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

func (*NOOP) EncryptMessage(message []byte) (crypt []byte, nonce []byte, err error) {
	return message, []byte{}, nil
}

func (*NOOP) DecryptMessage(cipher []byte, nonce []byte) (message []byte, err error) {
	return cipher, nil
}

type basicEncrypter struct {
	hasher             crypto.Hash
	senderPrivateKey   *rsa.PrivateKey
	recipientPublicKey *rsa.PublicKey
	label              []byte
}

type basicDecrypter struct {
	hasher              crypto.Hash
	recipientPrivateKey *rsa.PrivateKey
	senderPublicKey     *rsa.PublicKey
	label               []byte
}

func NewBasicEncrypter(hash crypto.Hash, senderPrivateKey *rsa.PrivateKey, recipientPublicKey *rsa.PublicKey) Encrypt {
	return &basicEncrypter{
		hasher:             hash,
		senderPrivateKey:   senderPrivateKey,
		recipientPublicKey: recipientPublicKey,
		label:              []byte("codex-basic-encrypter"),
	}
}

func NewBasicDecrypter(hash crypto.Hash, recipientPrivateKey *rsa.PrivateKey, senderPublicKey *rsa.PublicKey) Decrypt {
	return &basicDecrypter{
		hasher:              hash,
		recipientPrivateKey: recipientPrivateKey,
		senderPublicKey:     senderPublicKey,
		label:               []byte("codex-basic-encrypter"),
	}
}

func (c *basicEncrypter) EncryptMessage(message []byte) ([]byte, []byte, error) {
	cipherdata, err := rsa.EncryptOAEP(
		c.hasher.New(),
		crypto_rand.Reader,
		c.recipientPublicKey,
		message,
		c.label,
	)
	if err != nil {
		return []byte(""), []byte{}, emperror.Wrap(err, "failed to encrypt message")
	}

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

	pssh := c.hasher.New()
	pssh.Write(message)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPSS(crypto_rand.Reader, c.senderPrivateKey, c.hasher, hashed, &opts)
	if err != nil {
		return []byte(""), []byte{}, emperror.Wrap(err, "failed to sign message")
	}

	return cipherdata, signature, nil
}

func (c *basicDecrypter) DecryptMessage(cipher []byte, nonce []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(
		c.hasher.New(),
		crypto_rand.Reader,
		c.recipientPrivateKey,
		cipher,
		c.label,
	)
	if err != nil {
		return []byte{}, emperror.Wrap(err, "failed to decrypt message")
	}

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

	pssh := c.hasher.New()
	pssh.Write(decrypted)
	hashed := pssh.Sum(nil)

	err = rsa.VerifyPSS(c.senderPublicKey, c.hasher, hashed, nonce, &opts)
	if err != nil {
		return []byte{}, emperror.Wrap(err, "failed to validate signature")
	}

	return decrypted, nil
}

type encryptBox struct {
	senderPrivateKey   [32]byte
	recipientPublicKey [32]byte
	sharedEncryptKey   *[32]byte
}

func NewBoxEncrypter(senderPrivateKey [32]byte, recipientPublicKey [32]byte) Encrypt {

	encrypter := encryptBox{
		senderPrivateKey:   senderPrivateKey,
		recipientPublicKey: recipientPublicKey,
		sharedEncryptKey:   new([32]byte),
	}

	box.Precompute(encrypter.sharedEncryptKey, &encrypter.recipientPublicKey, &encrypter.senderPrivateKey)

	return &encrypter
}

func (enBox *encryptBox) EncryptMessage(message []byte) ([]byte, []byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return []byte(""), []byte{}, emperror.Wrap(err, "failed to generate nonce")
	}

	encrypted := box.SealAfterPrecomputation(nil, message, &nonce, enBox.sharedEncryptKey)

	return encrypted, nonce[:], nil
}

type decryptBox struct {
	recipientPrivateKey [32]byte
	senderPublicKey     [32]byte
	sharedDecryptKey    *[32]byte
}

func NewBoxDecrypter(recipientPrivateKey [32]byte, senderPublicKey [32]byte) Decrypt {

	decrypter := decryptBox{
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
