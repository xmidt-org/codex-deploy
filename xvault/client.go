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

package xvault

import (
	"github.com/hashicorp/vault/api"
)

type (
	reader interface {
		read(key string) (map[string]interface{}, error)
	}
	authenticator interface {
		getAuth(path string, data map[string]interface{}) (*secretAuth, error)
		setToken(token string)
	}
)

type secretAuth struct {
	*api.SecretAuth
}

type clientDecorator struct {
	*api.Client
}

func (c *clientDecorator) read(key string) (map[string]interface{}, error) {
	secretValues, err := c.Logical().Read(key)
	if err != nil {
		return nil, err
	}
	return secretValues.Data, nil
}

func (c *clientDecorator) getAuth(path string, data map[string]interface{}) (*secretAuth, error) {
	resp, err := c.Logical().Write(path, data)
	if err != nil {
		return nil, err
	}
	return &secretAuth{resp.Auth}, nil
}

func (c *clientDecorator) setToken(token string) {
	return
}

func newClient(config *api.Config) (*clientDecorator, error) {
	client, err := api.NewClient(config)
	return &clientDecorator{client}, err
}
