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
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

var (
	ErrEmptyRoleSecretID = errors.New("RoleID and SecretID can't be empty")
	ErrEmptyBasePath     = errors.New("BasePath can't be empty")
)

type Config struct {
	Address    string
	RoleID     string
	SecretID   string
	MaxRetries int
	BasePath   string
}

type Client struct {
	client   reader
	basePath string
}

func Initialize(v *viper.Viper) (*Client, error) {
	c := &Config{}
	v.UnmarshalKey("vault", c)
	return initialize(*c)
}

func initialize(config Config) (*Client, error) {
	if config.RoleID == "" || config.SecretID == "" {
		return nil, ErrEmptyRoleSecretID
	}
	if config.BasePath == "" {
		return nil, ErrEmptyBasePath

	}
	conf := &api.Config{
		Address: "",
	}

	if config.Address != "" {
		conf.Address = config.Address
	}
	if config.MaxRetries != 0 {
		conf.MaxRetries = config.MaxRetries
	}

	client, err := newClient(conf)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"role_id":   config.RoleID,
		"secret_id": config.SecretID,
	}

	err = authenticate(client, "auth/approle/login", data)

	return &Client{
		client:   client,
		basePath: config.BasePath,
	}, nil
}

func authenticate(auth authenticator, path string, data map[string]interface{}) error {
	resp, err := auth.getAuth(path, data)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("no auth info returned")
	}

	auth.setToken(resp.ClientToken)
	return nil
}

func (c *Client) GetUsernamePassword(stage string, key string) (string, string) {
	if key == "" {
		return "", ""
	}
	var path = c.basePath
	if stage != "" {
		path += fmt.Sprintf("/%s/%s", stage, key)
	} else {
		path += fmt.Sprintf("/%s", key)
	}
	data, err := c.client.read(path)
	if err != nil {
		return "", ""
	}
	var (
		usr string
		pwd string
	)
	if result, ok := data["usr"].(string); ok {
		usr = result
	}
	if result, ok := data["pwd"].(string); ok {
		pwd = result
	}

	return usr, pwd
}
