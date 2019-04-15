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

package main

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/nacl/box"
	"io/ioutil"
	"os"
)

var (
	privatePath string
	publicPath  string
)

func init() {
	flag.StringVar(&privatePath, "private", "private.pem", "output path for private key")
	flag.StringVar(&publicPath, "public", "public.pem", "output path for public key")
}

func createBoxFiles(args []string) int {
	flag.Parse()

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate random key %s\n", err.Error())
		return 1
	}

	privateData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "BOX PRIVATE KEY",
			Bytes: privateKey[:],
		},
	)

	publicData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "BOX PUBLIC KEY",
			Bytes: publicKey[:],
		},
	)

	err = ioutil.WriteFile(privatePath, privateData, 0400)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to file %s\n", err.Error())
		return 1
	}
	err = ioutil.WriteFile(publicPath, publicData, 0400)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to file %s\n", err.Error())
		return 1
	}
	return 0
}

func main() {
	os.Exit(createBoxFiles(os.Args))
}
