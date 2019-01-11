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
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

type Config struct {
	Address    string
	RoleID     string
	SecretID   string
	MaxRetries int
	BasePath   string
}

type Client struct {
	client   *api.Client
	basePath string
}

func Initialize(v *viper.Viper) (*Client, error) {
	c := &Config{}
	v.UnmarshalKey("vault", c)
	return newClient(*c)
}

func newClient(config Config) (*Client, error) {
	if config.RoleID == "" || config.SecretID == "" {
		return nil, fmt.Errorf("RoleID and SecretID can't be empty")
	}
	if config.BasePath == "" {
		return nil, fmt.Errorf("BasePath can't be empty")

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

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"role_id":   config.RoleID,
		"secret_id": config.SecretID,
	}

	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return nil, err
	}
	if resp.Auth == nil {
		return nil, fmt.Errorf("no auth info returned")
	}

	client.SetToken(resp.Auth.ClientToken)
	return &Client{
		client:   client,
		basePath: config.BasePath,
	}, nil
}

func (c *Client) GetKey(key string) (map[string]interface{}, error) {
	secretValues, err := c.client.Logical().Read(key)
	if err != nil {
		return nil, err
	}
	return secretValues.Data, nil
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
	data, err := c.GetKey(path)
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
