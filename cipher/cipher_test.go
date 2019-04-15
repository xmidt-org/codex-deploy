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
	crypto_rand "crypto/rand" // Custom so it's clear which rand we're using.
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"
	"testing"
)

func TestBasicCrypt(t *testing.T) {

	tests := []struct {
		size        int
		description string
		hashAlgo    crypto.Hash
		errOnLarge  bool
	}{
		{2048, "basic key", crypto.SHA256, true},
		{4096, "large key", crypto.BLAKE2b_512, true},
		{2048, "basic key with SHA512", crypto.SHA512, true},
		{1024, "basic key with MD5", crypto.MD5, true},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			t.Log(tc.size)
			require := require.New(t)

			senderPrivateKey := GeneratePrivateKey(tc.size)
			require.NotNil(senderPrivateKey)

			recipientPrivateKey := GeneratePrivateKey(tc.size)
			require.NotNil(recipientPrivateKey)

			encrypter := NewBasicEncrypter(tc.hashAlgo, senderPrivateKey, &recipientPrivateKey.PublicKey)
			require.NotEmpty(encrypter)
			decrypter := NewBasicDecrypter(tc.hashAlgo, recipientPrivateKey, &senderPrivateKey.PublicKey)
			require.NotEmpty(decrypter)

			testCryptoPair(t, encrypter, decrypter, tc.errOnLarge)
		})
	}
}

func testCryptoPair(t *testing.T, encrypter Encrypt, decrypter Decrypt, errOnLarge bool) {

	tests := []struct {
		str         string
		description string
	}{
		{"Hello World", "basic"},
		{"Hello, 世界", "complex characters"},
		{"Half-giant jinxes peg-leg gillywater broken glasses large black dog Great Hall. Nearly-Headless Nick now" +
			" string them together, and answer me this, which creature would you be unwilling to kiss? Poltergeist sticking" +
			" charm, troll umbrella stand flying cars golden locket Lily Potter. Pumpkin juice Trevor wave your wand out" +
			" glass orbs, a Grim knitted hats. Stan Shunpike doe patronus, suck his soul Muggle-Born large order of drills" +
			" the trace. Bred in captivity fell through the veil, quaffle blue flame ickle diddykins Aragog. Yer a wizard," +
			" Harry Doxycide the woes of Mrs. Weasley Goblet of Fire." +
			"refect’s bathroom Trelawney veela squashy armchairs, SPEW: Gamp’s Elemental Law of Transfiguration. Magic" +
			" Nagini bezoar, Hippogriffs Headless Hunt giant squid petrified. Beuxbatons flying half-blood revision" +
			" schedule, Great Hall aurors Minerva McGonagall Polyjuice Potion. Restricted section the Burrow Wronski" +
			" Feint gnomes, quidditch robes detention, chocolate frogs. Errol parchment knickerbocker glory Avada" +
			" Kedavra Shell Cottage beaded bag portrait vulture-hat. Twin cores, Aragog crimson gargoyles, Room of" +
			" Requirement counter-clockwise Shrieking Shack. Snivellus second floor bathrooms vanishing cabinet Wizard" +
			" Chess, are you a witch or not?", "large string"},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			message := []byte(tc.str)
			encodedMSG, nonce, err := encrypter.EncryptMessage(message)
			if !errOnLarge {
				require.NoError(err)
			} else if err != nil {
				return
			}
			assert.NoError(err)
			assert.NotEmpty(encodedMSG)

			msg, err := decrypter.DecryptMessage(encodedMSG, nonce)

			assert.NoError(err)
			assert.Equal(message, msg)
		})
	}
}

func TestBoxCipher(t *testing.T) {
	require := require.New(t)

	senderPublicKey, senderPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	require.NoError(err)

	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	require.NoError(err)

	encrypter := NewBoxEncrypter(*senderPrivateKey, *recipientPublicKey)
	require.NotEmpty(encrypter)
	decrypter := NewBoxDecrypter(*recipientPrivateKey, *senderPublicKey)
	require.NotEmpty(decrypter)

	testCryptoPair(t, encrypter, decrypter, false)
}
