// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"fmt"
	"net/url"
)

type API interface {
	WithRegistry(*url.URL) API
	HasTag(string, string) (bool, error)
}

// Client is a client for the registry API.
type Client struct {
	registry string
	Connector
}

// NewClient creates a new registry client.
func NewClient(connector Connector) API {
	return &Client{
		Connector: connector,
	}
}

// WithRegistry creates a new client with a registry.
func (c *Client) WithRegistry(reg *url.URL) API {
	c.registry = reg.String()
	return c
}

// HasTag checks if a tag exists in a registry.
func (c *Client) HasTag(name, tag string) (bool, error) {
	url := fmt.Sprintf("%s/v2/%s/tags/list", c.registry, name)
	var items TagsList
	err := c.Get(url, &items)
	if err != nil {
		return false, err
	}

	for _, t := range items.Tags {
		if t == tag {
			return true, nil
		}
	}

	return false, nil
}

var _ API = &Client{}
