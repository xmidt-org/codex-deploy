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

// KeyType is an enum for how the key can be used.
type KeyType string

const (
	PublicKey           KeyType = "publicKey"
	PrivateKey          KeyType = "privateKey"
	SenderPrivateKey    KeyType = "senderPrivateKey"
	SenderPublicKey     KeyType = "senderPublicKey"
	RecipientPrivateKey KeyType = "recipientPrivateKey"
	RecipientPublicKey  KeyType = "recipientPublicKey"
)

func hasBothEncryptKeys(data map[KeyType]string) bool {
	_, privateOK := data[SenderPrivateKey]
	_, publicOK := data[RecipientPublicKey]
	return privateOK && publicOK
}

func hasBothDecryptKeys(data map[KeyType]string) bool {
	_, privateOK := data[RecipientPrivateKey]
	_, publicOK := data[SenderPublicKey]
	return privateOK && publicOK
}
