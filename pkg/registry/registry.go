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

import "fmt"

type API interface {
	HasTag(string, string, string) (bool, error)
}

type Client struct {
	Connector
}

func NewClient(connector Connector) *Client {
	return &Client{
		Connector: connector,
	}
}

func (c *Client) HasTag(reg, name, tag string) (bool, error) {
	url := fmt.Sprintf("%s/v2/%s/tags/list", reg, name)
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
