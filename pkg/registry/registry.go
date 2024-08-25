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
