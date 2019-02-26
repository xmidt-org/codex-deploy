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
	"golang.org/x/crypto/blake2b"
	"hash"
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

// I the Interface for cipher to encrypt decrypt and verify the message
type Interface interface {
	Crypt
	Verify
}

// Crypt represents the ability to encrypt and decrypt messages
type Crypt interface {
	// EncryptMessage attempts to encode the message into an array of bytes.
	// and error will be returned if failed to encode the message.
	EncryptMessage(message []byte) ([]byte, error)

	// DecryptMessage attempts to decode the message into a string.
	// and error will be returned if failed to decode the message.
	DecryptMessage(cipher []byte) ([]byte, error)
}

// Verify is used to sign and verify the signature of a message
type Verify interface {
	// Sign attempts to sign the message into an array of bytes
	// and an error will be returned if a failure is encountered while signing the message.
	Sign(message []byte) ([]byte, error)

	// VerifyMessage will return true if the message was successfully verified
	VerifyMessage(message []byte, signature []byte) bool
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

// NOOP will just return the message
type NOOP struct{}

func (noop *NOOP) EncryptMessage(message []byte) ([]byte, error) {
	return message, nil
}

func (noop *NOOP) DecryptMessage(cipher []byte) ([]byte, error) {
	return cipher, nil
}

func (noop *NOOP) Sign(message []byte) ([]byte, error) {
	return message, nil
}

func (noop *NOOP) VerifyMessage(message []byte, signature []byte) bool {
	if (message == nil) != (signature == nil) {
		return false
	}

	if len(message) != len(signature) {
		return false
	}

	for i := range message {
		if message[i] != signature[i] {
			return false
		}
	}

	return true
}

type crypter struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	hasher     crypto.Hash
	label      []byte
}

func NewCrypter(hash crypto.Hash, key *rsa.PrivateKey) Interface {
	return &crypter{
		privateKey: key,
		publicKey:  &key.PublicKey,
		hasher:     hash,
	}
}

func (c *crypter) EncryptMessage(message []byte) ([]byte, error) {
	cipherdata, err := rsa.EncryptOAEP(
		c.hasher.New(),
		rand.Reader,
		c.publicKey,
		message,
		c.label,
	)
	if err != nil {
		return []byte(""), emperror.Wrap(err, "failed to encrypt message")
	}

	return cipherdata, nil
}

func (c *crypter) DecryptMessage(cipher []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(
		c.hasher.New(),
		rand.Reader,
		c.privateKey,
		cipher,
		c.label,
	)
	if err != nil {
		return []byte{}, emperror.Wrap(err, "failed to decrypt message")
	}
	return decrypted, nil
}

func (c *crypter) Sign(message []byte) ([]byte, error) {
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

	pssh := c.hasher.New()
	pssh.Write(message)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, c.privateKey, c.hasher, hashed, &opts)
	if err != nil {
		return []byte{}, emperror.Wrap(err, "failed to sign message")
	}

	return signature, nil
}
func (c *crypter) VerifyMessage(message []byte, signature []byte) bool {
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example

	pssh := c.hasher.New()
	pssh.Write(message)
	hashed := pssh.Sum(nil)

	err := rsa.VerifyPSS(c.publicKey, c.hasher, hashed, signature, &opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from rsa.VerifyPSS: %s\n", err)

		return false
	}

	return true
}
